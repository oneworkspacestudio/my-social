// Generated from screens/Shell.aui; edit the AUI source.
import { Alert, Button, Heading, Icon, Image, Link, Page, Row, Section, Spacer, Spinner, Stack, Text } from "@/components/ui"

export function Shell({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Stack variant="sidebar">
        <Link href="#home" variant="brand">
          <Image src={data.logo} alt="My Social" width={42} height={42} />
          <Stack gap="xs">
            <Heading level={2}>My Social</Heading>
            <Text variant="eyebrow">ONE WORKSPACE</Text>
          </Stack>
        </Link>
        {data.nav.map((item, i) => (
          <Link key={i} href={item.href} variant={item.variant} aria-label={item.title}>
            <Icon name={item.icon} />
            <Text>{item.title}</Text>
          </Link>
        ))}
        <Spacer />
        <Button variant="create" onClick={() => onAction("create")}>
          <Icon name="Plus" />
          <Text>Create a moment</Text>
        </Button>
        <Link href="#settings" variant="nav">
          <Icon name="Settings2" />
          <Text>Settings</Text>
        </Link>
        {data.workspace && (
          <Link href={data.workspace} variant="workspace-link">← One Workspace</Link>
        )}
      </Stack>
      <Stack variant="workspace">
        <Row variant="topbar">
          <Stack gap="xs">
            <Text variant="eyebrow">YOUR MOMENTS. YOUR PEOPLE.</Text>
            <Heading level={1}>{data.heading}</Heading>
          </Stack>
          <Spacer />
          <Button variant="icon" onClick={() => onAction("theme")} aria-label="Toggle color theme"><Icon name="SunMoon" /></Button>
          <Button variant="icon" onClick={() => onAction("apps")} aria-label="Open Workspace apps"><Icon name="Grid2X2" /></Button>
          <Link href="#notifications" variant="icon" aria-label="Notifications"><Icon name="Bell" /></Link>
          <Link href="#settings" variant="avatar" aria-label="Account settings">{data.initials}</Link>
        </Row>
        <Stack variant="content">
          {data.offline && (
            <Alert tone="warning">You’re offline. Device drafts remain available. Publishing needs a connection.</Alert>
          )}
          {data.error && (
            <Alert tone="error">{data.error}</Alert>
          )}
          {data.notice && (
            <Alert tone="success">{data.notice}</Alert>
          )}
          {data.busy && (
            <Row>
              <Spinner />
              <Text>{data.busy}</Text>
            </Row>
          )}
          <Section id="screen-outlet" />
        </Stack>
        <Text variant="footer">{data.versionText}</Text>
      </Stack>
      <Row variant="bottomnav">
        {data.mobileNav.map((item, i) => (
          <Link key={i} href={item.href} variant={item.variant} aria-label={item.title}>
            <Icon name={item.icon} />
            <Text>{item.title}</Text>
          </Link>
        ))}
      </Row>
    </Page>
  )
}
