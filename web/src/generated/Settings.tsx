// Generated from screens/Settings.aui; edit the AUI source.
import { Button, Card, Checkbox, Field, Heading, Input, Page, Row, Stack, Text } from "@/components/ui"

export function Settings({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Stack variant="settings">
        <Card>
          <Heading level={2}>Your profile</Heading>
          <Field label="Handle"><Input name="handle" /></Field>
          <Field label="Bio"><Input name="bio" multiline={true} /></Field>
          <Checkbox name="private">Private account — approve new followers</Checkbox>
          <Button onClick={() => onAction("profileSave")}>Save profile</Button>
        </Card>
        <Card>
          <Heading level={2}>Notifications</Heading>
          <Text>Push requires HTTPS and browser support. On iPhone or iPad, install My Social on your Home Screen first.</Text>
          <Checkbox name="push">Enable private push notifications</Checkbox>
          <Row>
            <Field label="Quiet hours start (server time)"><Input name="quietStart" type="time" /></Field>
            <Field label="Quiet hours end (server time)"><Input name="quietEnd" type="time" /></Field>
          </Row>
          <Button onClick={() => onAction("pushSave")}>Save notifications</Button>
          <Button onClick={() => onAction("testPush")} variant="secondary">Send test notification</Button>
        </Card>
        <Card>
          <Heading level={2}>Install My Social</Heading>
          <Text>Add this app to your Home Screen for an app-like experience. On iOS, use Share → Add to Home Screen.</Text>
          <Button onClick={() => onAction("install")} variant="secondary">Install / show instructions</Button>
          <Button onClick={() => onAction("updateApp")} variant="secondary">Apply available update</Button>
        </Card>
        <Card>
          <Heading level={2}>People & privacy</Heading>
          {data.relations.map((item, i) => (
            <Row key={i}>
              <Text>{item.label}</Text>
              <Button onClick={() => onAction("approve")} item={item.actor} variant="secondary">Approve request</Button>
              <Button onClick={() => onAction("reject")} item={item.actor} variant="secondary">Remove / reject</Button>
            </Row>
          ))}
          <Button onClick={() => onAction("export")} variant="secondary">Export my account & originals</Button>
          <Button onClick={() => onAction("logout")} variant="secondary">Sign out</Button>
          <Button onClick={() => onAction("deleteAccount")} variant="danger">Delete My Social account</Button>
        </Card>
        {data.owner && (
          <Card>
            <Heading level={2}>Community moderation</Heading>
            {data.reports.map((item, i) => (
              <Row key={i}>
                <Text>{item.reason}</Text>
                <Button onClick={() => onAction("removeReport")} item={item.id} variant="danger">Remove content</Button>
                <Button onClick={() => onAction("dismissReport")} item={item.id} variant="secondary">Dismiss</Button>
              </Row>
            ))}
          </Card>
        )}
        <Card>
          <Heading level={2}>Connected services</Heading>
          <Text>{data.storageText}</Text>
          <Text>{data.videoText}</Text>
          <Text>Live, payments, licensed music, automatic captions and a My Chat messaging adapter are not available in this preview.</Text>
          <Text variant="caption">{data.identityText}</Text>
        </Card>
      </Stack>
    </Page>
  )
}
