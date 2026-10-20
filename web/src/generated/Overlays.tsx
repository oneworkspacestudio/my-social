// Generated from screens/Overlays.aui; edit the AUI source.
import { Badge, Button, Card, Field, Grid, Heading, Image, Input, Link, Page, Row, Section, Select, Spacer, Text } from "@/components/ui"

export function Overlays({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      {data.commentsOpen && (
        <Card variant="dialog" aria-label="Post comments">
          <Row>
            <Heading level={2}>The conversation</Heading>
            <Spacer />
            <Button onClick={() => onAction("closeModal")} variant="secondary">Close</Button>
          </Row>
          {data.comments.map((item, i) => (
            <Card key={i} variant="comment">
              <Text variant="handle">{item.profile.handle}</Text>
              <Text>{item.body}</Text>
              <Text variant="caption">{item.date}</Text>
              <Button onClick={() => onAction("reply")} item={item.id} variant="text">Reply</Button>
            </Card>
          ))}
          <Field label={data.replyLabel}><Input name="comment" multiline={true} placeholder="Add a thoughtful comment…" /></Field>
          <Button onClick={() => onAction("sendComment")}>Send comment</Button>
        </Card>
      )}
      {data.previewOpen && (
        <Card variant="dialog" aria-label="Publication preview">
          <Row>
            <Heading level={2}>Your moment, as they’ll see it</Heading>
            <Spacer />
            <Button onClick={() => onAction("closeModal")} variant="secondary">Close</Button>
          </Row>
          <Section id="preview-media" />
          <Text>{data.caption}</Text>
          <Badge tone="muted">{data.visibility}</Badge>
          <Button onClick={() => onAction("publish")}>Publish moment</Button>
        </Card>
      )}
      {data.storyOpen && (
        <Card variant="story-dialog" aria-label="Story viewer">
          <Row>
            <Text variant="handle">{data.storyHandle}</Text>
            <Spacer />
            <Button onClick={() => onAction("closeModal")} variant="secondary">Close</Button>
          </Row>
          <Section id="story-media" />
          <Section id="story-progress" />
          <Text>{data.storyCaption}</Text>
          {data.storyPoll && (
            <>
              <Heading level={3}>{data.storyPoll}</Heading>
              {data.storyChoices.map((item, i) => (
                <Button key={i} onClick={() => onAction("vote")} item={item.value} variant="secondary">{item.value}</Button>
              ))}
            </>
          )}
          {data.storyLink && (
            <Link href={data.storyLink} external={true}>Open Story link ↗</Link>
          )}
          <Row>
            <Button onClick={() => onAction("storyPrevious")} variant="secondary">Previous</Button>
            <Button onClick={() => onAction("storyPause")} variant="secondary">{data.pauseLabel}</Button>
            <Button onClick={() => onAction("storyNext")} variant="secondary">Next</Button>
          </Row>
          <Field label="Reply to Story"><Input name="storyReply" placeholder="Say something…" /></Field>
          <Button onClick={() => onAction("storyReply")}>Send reply</Button>
        </Card>
      )}
      {data.libraryOpen && (
        <Card variant="dialog" aria-label="My Photos picker">
          <Row>
            <Heading level={2}>Your My Social media in My Photos</Heading>
            <Spacer />
            <Button onClick={() => onAction("closeModal")} variant="secondary">Close</Button>
          </Row>
          <Text>Only your protected My Social originals appear here. Choose an ordinary My Photos file below to import with its existing permissions.</Text>
          <Grid columns={3}>
            {data.library.map((item, i) => (
              <Button key={i} onClick={() => onAction("pickMedia")} item={item.id} variant="media-pick"><Image src={item.url} alt="Your media" width={120} height={120} /></Button>
            ))}
          </Grid>
          {data.workspace && (
            <>
              <Field label="My Photos folder (inside your Drive scope)"><Input name="photosFolder" placeholder="/Photos" /></Field>
              <Button onClick={() => onAction("listPhotos")} variant="secondary">Browse ordinary My Photos</Button>
              {data.photoFiles.map((item, i) => (
                <Row key={i}>
                  <Text>{item.name}</Text>
                  <Button onClick={() => onAction("importPhoto")} item={item.path} variant="secondary">Import photo</Button>
                </Row>
              ))}
            </>
          )}
        </Card>
      )}
      {data.editOpen && (
        <Card variant="dialog" aria-label="Edit post details">
          <Heading level={2}>Edit your moment</Heading>
          <Field label="Caption"><Input name="editCaption" multiline={true} /></Field>
          <Field label="Alt text"><Input name="editAlt" /></Field>
          <Field label="Audience"><Select name="editVisibility" options={data.audienceOptions} /></Field>
          <Row>
            <Button onClick={() => onAction("savePost")}>Save changes</Button>
            <Button onClick={() => onAction("closeModal")} variant="secondary">Close</Button>
          </Row>
        </Card>
      )}
    </Page>
  )
}
