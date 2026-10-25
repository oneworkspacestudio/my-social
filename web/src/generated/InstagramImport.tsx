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
          <Text variant="caption">JPEG, PNG, GIF, WebP, MP4 and WebM · 100 MiB per file · imported videos up to 10 minutes need FFmpeg and FFprobe.</Text>
          <Text variant="caption">Stories and archived posts stay archived in Creator Studio. Followers, messages, likes, comments and profile details are not imported.</Text>
          <Button onClick={() => onAction("native")}>Choose ZIP for background upload</Button>
          <Text variant="caption">Windows opens a file picker. Select your ZIP, then you can close this page while it uploads. Keep the computer on and the ZIP in place.</Text>
          {data.nativeLink && (
            <Link href={data.nativeLink}>Open Windows file picker</Link>
          )}
          <Link href={data.nativeDownload} external={true}>Install Windows uploader on this computer</Link>
          <Text variant="caption">The uploader is included on your Workspace PC. On another Windows computer, download and run it once, then return here.</Text>
          <Heading level={3}>Or upload in this tab</Heading>
          <Field label="Instagram export ZIP"><Input name="archive" type="file" accept=".zip,application/zip" /></Field>
          {data.filename && (
            <Text variant="caption">{data.filename}</Text>
          )}
          <Button onClick={() => onAction("preview")} disabled={data.previewDisabled}>{data.previewLabel}</Button>
          {data.serverMessage && (
            <Text>{data.serverMessage}</Text>
          )}
          {data.busy && (
            <>
              <Alert>{data.busy}</Alert>
              <Spinner variant="import-progress" value={data.progress} label={data.busy} />
              <Text variant="caption">{data.phaseCounts}</Text>
              <Text variant="caption">{data.elapsedText}</Text>
            </>
          )}
          {data.showPause && (
            <Button onClick={() => onAction("pause")} cancel={true} variant="secondary">Pause upload</Button>
          )}
          {data.showCancel && (
            <Button onClick={() => onAction("cancel")} cancel={true} variant="secondary">Cancel import</Button>
          )}
          {data.showRetry && (
            <Button onClick={() => onAction("retry")} cancel={true} variant="secondary">Retry server processing</Button>
          )}
          {data.error && (
            <Alert tone="error">{data.error}</Alert>
          )}
          {data.notice && (
            <Alert tone="success">{data.notice}</Alert>
          )}
        </Card>
        {data.hasJobs && (
          <Card>
            <Heading level={2}>Recent imports</Heading>
            {data.jobs.map((item, i) => (
              <Button key={i} onClick={() => onAction("openJob")} item={item.id} variant="secondary">{item.label}</Button>
            ))}
          </Card>
        )}
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
