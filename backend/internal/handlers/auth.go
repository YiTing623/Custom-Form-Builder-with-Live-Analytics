package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/google/uuid"

	"github.com/YiTing623/Custom-Form-Builder-with-Live-Analytics/internal/db"
	"github.com/YiTing623/Custom-Form-Builder-with-Live-Analytics/internal/models"
)

type AuthHandler struct {
	Store     *db.MongoStore
	JWTSecret []byte
}

func NewAuthHandler(s *db.MongoStore, secret []byte) *AuthHandler {
	return &AuthHandler{Store: s, JWTSecret: secret}
}

type registerReq struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}
type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type authResp struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var in registerReq
	if err := c.BodyParser(&in); err != nil || in.Email == "" || len(in.Password) < 6 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	pw, _ := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	u := models.User{
		ID:           uuid.NewString(),
		Email:        in.Email,
		Name:         in.Name,
		PasswordHash: pw,
		Created:      time.Now().Unix(),
	}

	ctx, cancel := context.WithTimeout(c.Context(), 8*time.Second)
	defer cancel()

	if _, err := h.Store.Users.InsertOne(ctx, u); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "email already registered")
	}

	tok, _ := h.makeToken(u.ID)
	c.Cookie(&fiber.Cookie{Name: "token", Value: tok, HTTPOnly: true, SameSite: "Lax"})
	return c.JSON(authResp{Token: tok, User: u})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var in loginReq
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 8*time.Second)
	defer cancel()

	var u models.User
	if err := h.Store.Users.FindOne(ctx, bson.M{"email": in.Email}).Decode(&u); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(in.Password)); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}

	tok, _ := h.makeToken(u.ID)
	c.Cookie(&fiber.Cookie{Name: "token", Value: tok, HTTPOnly: true, SameSite: "Lax"})
	return c.JSON(authResp{Token: tok, User: u})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	uid, _ := c.Locals("userId").(string)
	if uid == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "not logged in")
	}
	ctx, cancel := context.WithTimeout(c.Context(), 8*time.Second)
	defer cancel()

	var u models.User
	if err := h.Store.Users.FindOne(ctx, bson.M{"_id": uid}).Decode(&u); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "not found")
	}
	u.PasswordHash = nil
	return c.JSON(u)
}

func (h *AuthHandler) makeToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(h.JWTSecret)
}
