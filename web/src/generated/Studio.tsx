// Generated from screens/Studio.aui; edit the AUI source.
import { Badge, Button, Card, Grid, Heading, Metric, Page, Row, Spacer, Stack, Text } from "@/components/ui"

export function Studio({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Stack>
        <Row>
          <Stack gap="xs">
            <Heading level={2}>Your creative space</Heading>
            <Text>Pick up an idea, see what’s scheduled, or revisit a moment.</Text>
          </Stack>
          <Spacer />
          <Button onClick={() => onAction("create")}>Create</Button>
        </Row>
        <Grid columns={3}>
          <Card>
            <Text variant="eyebrow">PUBLISHED MOMENTS</Text>
            <Metric value={data.publishedCount} />
          </Card>
          <Card>
            <Text variant="eyebrow">LIKES RECEIVED</Text>
            <Metric value={data.likesCount} />
          </Card>
          <Card>
            <Text variant="eyebrow">YOUR DRAFTS</Text>
            <Metric value={data.draftCount} />
          </Card>
        </Grid>
        <Heading level={2}>Drafts</Heading>
        {data.drafts.map((item, i) => (
          <Card key={i} variant="studio-row">
            <Text>{item.title}</Text>
            <Text variant="caption">{item.date}</Text>
            <Spacer />
            <Button onClick={() => onAction("resumeDraft")} item={item.id}>Continue editing</Button>
            <Button onClick={() => onAction("deleteDraft")} item={item.id} variant="secondary">Delete draft</Button>
          </Card>
        ))}
        <Heading level={2}>Published, scheduled & archived</Heading>
        {data.posts.map((item, i) => (
          <Card key={i} variant="studio-row">
            <Text>{item.caption}</Text>
            <Badge tone="muted">{item.state}</Badge>
            <Text variant="caption">{item.date}</Text>
            <Spacer />
            <Button onClick={() => onAction("openPost")} item={item.id} variant="secondary">View</Button>
            <Button onClick={() => onAction("restorePost")} item={item.id} variant="secondary">Restore</Button>
            <Button onClick={() => onAction("addHighlight")} item={item.id} variant="secondary">Add to Highlight</Button>
            <Button onClick={() => onAction("deletePost")} item={item.id} variant="danger">Delete</Button>
          </Card>
        ))}
        <Heading level={2}>Video processing</Heading>
        {data.jobs.map((item, i) => (
          <Card key={i} variant="studio-row">
            <Text>{item.state}</Text>
            <Text>{item.progressText}</Text>
            <Text>{item.error}</Text>
          </Card>
        ))}
        {data.empty && (
          <Card variant="empty">
            <Heading level={3}>Every moment starts with an idea.</Heading>
            <Text>Your drafts and published work will appear here.</Text>
          </Card>
        )}
      </Stack>
    </Page>
  )
}
