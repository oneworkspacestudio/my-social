// Generated from screens/InstagramImport.aui; edit the AUI source.
import { Alert, Badge, Button, Card, Field, Heading, Input, Link, Page, Row, Select, Spinner, Stack, Text } from "@/components/ui"

export function InstagramImport({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Stack variant="instagram-import">
        <Card>
          <Heading level={2}>Bring your Instagram moments</Heading>
          <Text>Import your photos, videos, captions, original dates and carousels into My Social.</Text>
          <Link href="https://github.com/oneworkspacestudio/my-social/releases/download/v0.1.0-preview.6-capture.2/my-social-instagram-capture-0.1.2.zip" external={true}>Download Chrome capture extension</Link>
          <Text variant="caption">The extension can capture accessible posts from an open Instagram tab and create import ZIPs. Extract it, then use Chrome → Extensions → Developer mode → Load unpacked.</Text>
          <Text>In Instagram, open Accounts Center → Your information and permissions → Export your information. Choose JSON, include posts and media, and download the ZIP.</Text>
          <Text variant="caption">ZIP up to 512 MiB · 1,000 posts per ZIP · import all available posts with one click. For larger exports, choose a smaller date range.</Text>
          <Text variant="caption">JPEG, PNG, GIF, WebP, MP4 and WebM · 100 MiB per file · videos up to 90 seconds need FFmpeg and FFprobe.</Text>
          <Text variant="caption">Stories and archived posts stay archived in Creator Studio. Followers, messages, likes, comments and profile details are not imported.</Text>
          <Field label="Instagram export ZIP"><Input name="archive" type="file" accept=".zip,application/zip" /></Field>
          {data.filename && (
            <Text variant="caption">{data.filename}</Text>
          )}
          <Button onClick={() => onAction("preview")} disabled={data.previewDisabled}>Preview ZIP</Button>
          {data.busy && (
            <>
              <Alert>{data.busy}</Alert>
              <Spinner variant="import-progress" value={data.progress} label={data.busy} />
              <Text variant="caption">{data.elapsedText}</Text>
              <Button onClick={() => onAction("stop")} cancel={true} variant="secondary">Stop</Button>
            </>
          )}
          {data.error && (
            <Alert tone="error">{data.error}</Alert>
          )}
          {data.notice && (
            <Alert tone="success">{data.notice}</Alert>
          )}
        </Card>
        {data.review && (
          <Card>
            <Heading level={2}>Review your import</Heading>
            <Text>{data.summary}</Text>
            <Text>A small sample is shown below. Import brings across every available post in this ZIP. Unavailable files and existing imports are skipped.</Text>
            <Field label="Who can see imported posts?"><Select name="visibility" options={data.audiences} /></Field>
            <Text variant="caption">Only me is selected by default. Originals stay protected. Existing imports keep their audience and are skipped on retry.</Text>
            <Text variant="caption">{data.sampleText}</Text>
            <Button onClick={() => onAction("apply")} disabled={data.applyDisabled}>{data.applyLabel}</Button>
            {data.items.map((item, i) => (
              <Card key={i} variant="import-item">
                <Row variant="import-actions">
                  <Badge>{item.kind}</Badge>
                  <Text variant="caption">{item.date}</Text>
                  <Badge tone={item.tone}>{item.status}</Badge>
                  <Text variant="caption">{item.archiveText}</Text>
                </Row>
                <Text variant="post-caption">{item.caption}</Text>
                <Text variant="caption">{item.mediaText}</Text>
                {item.reason && (
                  <Text variant="caption">{item.reason}</Text>
                )}
              </Card>
            ))}
          </Card>
        )}
        <Row variant="import-actions">
          <Link href="#settings">Back to Settings</Link>
          <Link href="#studio">Open Creator Studio</Link>
        </Row>
      </Stack>
    </Page>
  )
}
