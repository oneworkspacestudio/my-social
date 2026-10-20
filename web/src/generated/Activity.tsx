// Generated from screens/Activity.aui; edit the AUI source.
import { Badge, Button, Card, Heading, Icon, Page, Row, Spacer, Stack, Text } from "@/components/ui"

export function Activity({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Stack variant="activity">
        <Row>
          <Heading level={2}>A little hello from your people</Heading>
          <Spacer />
          <Button onClick={() => onAction("readAll")} variant="secondary">Mark all read</Button>
        </Row>
        {data.events.map((item, i) => (
          <Card key={i} variant="activity-row">
            <Icon name="Bell" />
            <Stack gap="xs">
              <Text>{item.label}</Text>
              <Text variant="caption">{item.date}</Text>
            </Stack>
            <Spacer />
            <Badge tone={item.tone}>{item.status}</Badge>
            <Button onClick={() => onAction("notification")} item={item.id} variant="secondary">Open</Button>
          </Card>
        ))}
        {data.empty && (
          <Card variant="empty">
            <Icon name="Heart" size={40} />
            <Heading level={2}>You’re all caught up.</Heading>
            <Text>Likes, follows, comments and Story responses will arrive here.</Text>
          </Card>
        )}
      </Stack>
    </Page>
  )
}
