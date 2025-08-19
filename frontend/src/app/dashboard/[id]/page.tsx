import DashboardClient from './DashboardClient';

export default async function Page({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <DashboardClient id={id} />;
}
