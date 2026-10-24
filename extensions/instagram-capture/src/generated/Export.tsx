// Generated from screens/Export.aui; edit the AUI source.
import { Alert, Badge, Button, Card, Checkbox, Field, Heading, Image, Input, Link, Page, Row, Select, Stack, Text } from "@/components/ui"

export function Export({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Stack>
        <Row>
          <Image src="icons/48.png" alt="My Social" width={48} height={48} />
          <Stack>
            <Text variant="eyebrow">INSTAGRAM → YOUR COMPUTER → MY SOCIAL</Text>
            <Heading level={1}>Download my Instagram uploads</Heading>
          </Stack>
        </Row>
        <Card>
          <Heading level={2}>1. Find my photos and videos</Heading>
          <Text>Keep your own Instagram profile open. This scans its uploads, including carousel photos and videos. Feed, saved, tagged and other accounts’ posts are excluded.</Text>
          <Field label="My Instagram username"><Input name="username" placeholder="Your Instagram username" /></Field>
          <Text variant="caption">{data.source}</Text>
          <Checkbox name="scroll">Scroll my profile to find older uploads</Checkbox>
          <Field label="Maximum scroll steps per grid"><Select name="scrolls" options={data.scrollOptions} /></Field>
          <Checkbox name="reels">Scan my Reels grid too</Checkbox>
          <Checkbox name="enrich">Open each discovered upload to find its photos, videos and carousel slides</Checkbox>
          <Text variant="caption">Scanning finds media first. Click Download below to save the actual files. Scanning a large profile can take several minutes; keep this tab open.</Text>
          <Row>
            <Button onClick={() => onAction("capture")} disabled={data.noSource}>Scan my uploads</Button>
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
            <Heading level={2}>2. Download my media</Heading>
            <Text>{data.summary}</Text>
            <Row>
              <Button onClick={() => onAction("all")} variant="secondary">Select all uploads</Button>
              <Button onClick={() => onAction("none")} variant="secondary">Clear selection</Button>
              <Button onClick={() => onAction("export")} disabled={data.noSelection}>Download my media ZIPs</Button>
            </Row>
            <Text>Your photos and videos go into Chrome Downloads → MySocial. Unzip a downloaded file and open media/posts/ to see them. Each ZIP also works with My Social’s Instagram importer.</Text>
            <Text variant="caption">A profile scan is best effort, not proof that every upload is backed up. Incomplete media is marked below. Stories/archives and unavailable uploads need Instagram’s account export. Videos over 90 seconds can be downloaded but cannot be imported unchanged into My Social.</Text>
            {data.posts.map((item, i) => (
              <Card key={i} variant="post">
                {item.hasPreview && (
                  <Image variant="preview" src={item.preview} alt="Upload preview" />
                )}
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
                <Link href={item.permalink} external={true}>View my upload on Instagram</Link>
                <Button onClick={() => onAction("toggle")} item={item.permalink} variant={item.variant} aria-pressed={item.selected}>{item.label}</Button>
              </Card>
            ))}
          </Card>
        )}
        {data.hasDownloads && (
          <Card>
            <Heading level={2}>My downloaded photos and videos</Heading>
            <Text>Saved in Chrome Downloads → MySocial. The real media files are inside each ZIP under media/posts/. Partial recoveries are under capture/incomplete/; capture/report.json lists any failures.</Text>
            <Button onClick={() => onAction("reveal")} variant="secondary">Show downloaded ZIP in folder</Button>
            {data.downloads.map((item, i) => (
              <Text key={i}>{item.label}</Text>
            ))}
          </Card>
        )}
        {data.hasWarnings && (
          <Card>
            <Heading level={2}>Missing media and scan notes</Heading>
            {data.warnings.map((item, i) => (
              <Text key={i}>{item.text}</Text>
            ))}
          </Card>
        )}
        <Text variant="caption">Nothing is uploaded by this extension. It uses your open Instagram session and normal pages; no password, cookie or private API export.</Text>
      </Stack>
    </Page>
  )
}
