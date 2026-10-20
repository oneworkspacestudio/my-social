// Generated from screens/Create.aui; edit the AUI source.
import { Badge, Button, Card, Checkbox, Field, Grid, Heading, Icon, Input, Page, Row, Section, Select, Spacer, Stack, Text } from "@/components/ui"

export function Create({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Stack variant="creator">
        <Row variant="steps">
          <Badge tone="success">01 / CAPTURE</Badge>
          <Text>02 / MAKE IT YOURS</Text>
          <Text>03 / SHARE</Text>
        </Row>
        <Card>
          <Row>
            <Heading level={2}>What’s your moment?</Heading>
            <Spacer />
            <Button onClick={() => onAction("closeCreate")} variant="secondary">Close</Button>
          </Row>
          <Row variant="mode-tabs">
            <Button onClick={() => onAction("kindPost")} variant={data.postVariant}>Post</Button>
            <Button onClick={() => onAction("kindStory")} variant={data.storyVariant}>Story</Button>
            <Button onClick={() => onAction("kindReel")} variant={data.reelVariant}>Reel</Button>
            <Button onClick={() => onAction("kindLive")} variant="secondary">Live · unavailable</Button>
          </Row>
          <Row variant="source-actions">
            <Button onClick={() => onAction("camera")} variant="secondary">
              <Icon name="Camera" />
              <Text>Open camera</Text>
            </Button>
            <Field label="Choose photos or videos"><Input name="files" type="file" accept="image/jpeg,image/png,image/gif,video/mp4,video/webm" multiple={true} /></Field>
            <Button onClick={() => onAction("myPhotos")} variant="secondary">My Photos</Button>
          </Row>
          <Text variant="caption">Originals are stored privately in My Photos / My Social. Only the version you publish is shared.</Text>
          <Section id="camera" />
          <Section id="media-strip" />
        </Card>
        {data.hasMedia && (
          <>
            <Card variant="editor-card">
              <Row>
                <Heading level={2}>Make it yours</Heading>
                <Spacer />
                <Button onClick={() => onAction("undo")} variant="icon" aria-label="Undo"><Icon name="Undo2" /></Button>
                <Button onClick={() => onAction("redo")} variant="icon" aria-label="Redo"><Icon name="Redo2" /></Button>
              </Row>
              <Section id="editor" />
              {data.isVideo ? (
                <Section id="timeline" />
              ) : (
                <>
                  <Row variant="edit-tools">
                    <Button onClick={() => onAction("addText")} variant="secondary">Add text</Button>
                    <Button onClick={() => onAction("addSticker")} variant="secondary">Add sticker</Button>
                    <Button onClick={() => onAction("brush")} variant="secondary">Draw / select</Button>
                    <Button onClick={() => onAction("deleteLayer")} variant="secondary">Delete layer</Button>
                    <Button onClick={() => onAction("layerUp")} variant="secondary">Bring forward</Button>
                    <Button onClick={() => onAction("rotate")} variant="secondary">Rotate</Button>
                    <Button onClick={() => onAction("flip")} variant="secondary">Flip</Button>
                    <Button onClick={() => onAction("removeImage")} variant="secondary">Remove image</Button>
                  </Row>
                  <Grid columns={3} variant="controls">
                    <Field label="Aspect ratio"><Select name="aspect" options={data.aspectOptions} /></Field>
                    <Field label="Crop zoom"><Input name="cropZoom" type="range" min={1} max={3} step={0.01} /></Field>
                    <Field label="Straighten"><Input name="straighten" type="range" min={-15} max={15} step={0.5} /></Field>
                    <Field label="Brightness"><Input name="brightness" type="range" min={-0.5} max={0.5} step={0.01} /></Field>
                    <Field label="Contrast"><Input name="contrast" type="range" min={-0.5} max={0.5} step={0.01} /></Field>
                    <Field label="Saturation"><Input name="saturation" type="range" min={-1} max={1} step={0.01} /></Field>
                    <Field label="Warmth"><Input name="warmth" type="range" min={-0.5} max={0.5} step={0.01} /></Field>
                    <Field label="Blur"><Input name="blur" type="range" min={0} max={0.3} step={0.01} /></Field>
                    <Field label="Filter"><Select name="filter" options={data.filterOptions} /></Field>
                    <Field label="Overlay text / emoji"><Input name="overlayText" multiline={true} placeholder="Write something… or add ✨" /></Field>
                    <Field label="Text size"><Input name="fontSize" type="number" min={12} max={160} /></Field>
                    <Field label="Text / brush color"><Input name="color" type="color" /></Field>
                    <Field label="Font"><Select name="font" options={data.fontOptions} /></Field>
                    <Field label="Brush width"><Input name="brushWidth" type="range" min={1} max={40} step={1} /></Field>
                    <Checkbox name="textBackground">Text highlight</Checkbox>
                  </Grid>
                  <Row>
                    <Button onClick={() => onAction("updateText")} variant="secondary">Apply text style</Button>
                    <Button onClick={() => onAction("template")} variant="secondary">Story template</Button>
                  </Row>
                  <Text variant="caption">Drag, resize and rotate overlays using the canvas handles. Double-click text to edit. Mouse wheel zooms the composition; touch controls support pinch zoom.</Text>
                </>
              )}
            </Card>
            <Card variant="publish-card">
              <Heading level={2}>Ready to share?</Heading>
              <Field label="Caption"><Input name="caption" multiline={true} placeholder="Tell the story behind it… #moments" /></Field>
              <Grid columns={2}>
                <Field label="Who can see this?"><Select name="visibility" options={data.audienceOptions} /></Field>
                <Field label="Alt text"><Input name="alt" placeholder="Describe the image for screen readers" /></Field>
                <Field label="Location (optional)"><Input name="location" placeholder="Add a place, never your GPS" /></Field>
                <Field label="Schedule (optional)"><Input name="schedule" type="datetime-local" /></Field>
              </Grid>
              {data.isStory && (
                <Row>
                  <Field label="Poll question (optional)"><Input name="poll" placeholder="What do you think?" /></Field>
                  <Field label="Poll choices (comma separated)"><Input name="pollChoices" placeholder="Love it, Tell me more" /></Field>
                  <Field label="Link sticker (HTTPS, optional)"><Input name="storyLink" type="url" /></Field>
                </Row>
              )}
              <Row variant="publish-actions">
                <Button onClick={() => onAction("preview")} variant="secondary">Preview</Button>
                <Button onClick={() => onAction("saveDraft")} variant="secondary">Save draft</Button>
                <Button onClick={() => onAction("publish")}>Publish moment</Button>
              </Row>
            </Card>
          </>
        )}
      </Stack>
    </Page>
  )
}
