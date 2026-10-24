// Generated from screens/Export.aui; edit the AUI source.
import { Alert, Badge, Button, Card, Checkbox, Field, Heading, Image, Link, Page, Row, Select, Stack, Text } from "@/components/ui"

export function Export({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Stack>
        <Row>
          <Image src="icons/48.png" alt="My Social" width={48} height={48} />
          <Stack>
            <Text variant="eyebrow">MY SOCIAL</Text>
            <Heading level={1}>Bring your Instagram moments</Heading>
          </Stack>
        </Row>
        <Card>
          <Heading level={2}>1. Capture your tab</Heading>
          <Text>Open a profile, feed, post or Reel on Instagram, then click this extension. Keep both tabs open during capture.</Text>
          <Text variant="caption">{data.source}</Text>
          <Checkbox name="scroll">Scroll the Instagram tab to load more posts</Checkbox>
          <Field label="Maximum scroll steps"><Select name="scrolls" options={data.scrollOptions} /></Field>
          <Checkbox name="enrich">Read discovered post pages for better captions, dates and carousels</Checkbox>
          <Text variant="caption">Only accessible page content is collected. No private API, password or cookie export. Messages and account/security pages are excluded.</Text>
          <Row>
            <Button onClick={() => onAction("capture")} disabled={data.noSource}>Capture content</Button>
            {data.busy && (
              <Button onClick={() => onAction("stop")} cancel={true} variant="secondary">Stop</Button>
            )}
          </Row>
          {data.busy && (
            <Alert>{data.busy}</Alert>
          )}
          {data.error && (
            <Alert tone="error">{data.error}</Alert>
          )}
          {data.notice && (
            <Alert>{data.notice}</Alert>
          )}
        </Card>
        {data.hasPosts && (
          <Card>
            <Heading level={2}>2. Choose your posts</Heading>
            <Text>{data.summary}</Text>
            <Row>
              <Button onClick={() => onAction("all")} variant="secondary">Select all</Button>
              <Button onClick={() => onAction("none")} variant="secondary">Clear selection</Button>
              <Button onClick={() => onAction("export")} disabled={data.noSelection}>Download import ZIPs</Button>
            </Row>
            <Text variant="caption">ZIP parts include their own media and manifest. Import each ZIP in My Social → Settings → Import Instagram ZIP. Captures are held in memory until this tab closes.</Text>
            <Text variant="caption">Missing dates use the capture date and are marked below. Larger carousels are split at 10 items. Videos over 90 seconds may be archived but cannot be imported unchanged into My Social.</Text>
            {data.posts.map((item, i) => (
              <Card key={i} variant="post">
                <Row>
                  <Badge>{item.kind}</Badge>
                  <Text variant="caption">{item.owner}</Text>
                  <Text variant="caption">{item.date}</Text>
                </Row>
                <Text>{item.caption}</Text>
                <Text variant="caption">{item.details}</Text>
                {item.warning && (
                  <Alert tone="warning">{item.warning}</Alert>
                )}
                <Link href={item.permalink} external={true}>View source post</Link>
                <Button onClick={() => onAction("toggle")} item={item.permalink} variant={item.variant} aria-pressed={item.selected}>{item.label}</Button>
              </Card>
            ))}
          </Card>
        )}
        {data.hasDownloads && (
          <Card>
            <Heading level={2}>Your downloaded ZIPs</Heading>
            {data.downloads.map((item, i) => (
              <Text key={i}>{item.label}</Text>
            ))}
          </Card>
        )}
        {data.hasWarnings && (
          <Card>
            <Heading level={2}>Capture notes</Heading>
            {data.warnings.map((item, i) => (
              <Text key={i}>{item.text}</Text>
            ))}
          </Card>
        )}
        <Text variant="caption">This is a best-effort capture of content available to your tab. Use Instagram’s official JSON export for the most complete account archive. Capture only content you own or have permission to copy. Nothing is uploaded by this extension.</Text>
      </Stack>
    </Page>
  )
}
