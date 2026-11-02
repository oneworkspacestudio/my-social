// Generated from screens/Create.aui; edit the AUI source.
import { Alert, Button, Checkbox, Field, Grid, Heading, Icon, Input, Page, Row, Section, Select, Spacer, Spinner, Stack, Text } from "@/components/ui"

export function Create({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Stack variant="creator">
        <Row variant="compose-header">
          <Button onClick={() => onAction("backCreate")} variant="icon" aria-label={data.backLabel}><Icon name={data.backIcon} /></Button>
          <Heading level={1}>{data.title}</Heading>
          <Spacer />
          {data.canNext ? (
            <Button onClick={() => onAction("nextCreate")} variant="next" disabled={data.nextDisabled}>Next</Button>
          ) : (
            <Button onClick={() => onAction("saveDraft")} variant="text">Save draft</Button>
          )}
        </Row>
        {data.selecting && (
          <Stack variant="compose-select">
            <Section id="selection-preview" />
            <Row variant="gallery-heading">
              <Heading level={2}>Photos and videos</Heading>
              <Spacer />
              <Button onClick={() => onAction("camera")} variant="icon" aria-label="Open camera"><Icon name="Camera" /></Button>
              <Button onClick={() => onAction("myPhotos")} variant="text">My Photos</Button>
            </Row>
            <Field label="Choose photos or videos"><Input name="files" type="file" accept="image/jpeg,image/png,image/gif,video/mp4,video/webm" multiple={true} /></Field>
            <Section id="camera" />
            <Section id="media-strip" />
            {data.hasMedia ? (
              <Row variant="selection-actions">
                <Text variant="caption">{data.selectionText}</Text>
                <Spacer />
                <Button onClick={() => onAction("removeImage")} variant="text">Remove selected</Button>
              </Row>
            ) : (
              <Text variant="caption">Choose up to 10 photos, or video clips for a Reel.</Text>
            )}
            <Row variant="compose-modes">
              <Button onClick={() => onAction("kindPost")} variant={data.postVariant}>Post</Button>
              <Button onClick={() => onAction("kindStory")} variant={data.storyVariant}>Story</Button>
              <Button onClick={() => onAction("kindReel")} variant={data.reelVariant}>Reel</Button>
            </Row>
          </Stack>
        )}
        {data.editing && (
          <Stack variant="compose-edit">
            <Section id="editor" />
            <Section id="media-strip" />
            {data.isVideo ? (
              <Section id="timeline" />
            ) : (
              <>
                <Row variant="compose-tools">
                  <Button onClick={() => onAction("toolCrop")} variant={data.cropVariant}>
                    <Icon name="Crop" />
                    <Text>Crop</Text>
                  </Button>
                  <Button onClick={() => onAction("toolFilter")} variant={data.filterVariant}>
                    <Icon name="SlidersHorizontal" />
                    <Text>Adjust</Text>
                  </Button>
                  <Button onClick={() => onAction("toolText")} variant={data.textVariant}>
                    <Icon name="Type" />
                    <Text>Text</Text>
                  </Button>
                  <Button onClick={() => onAction("toolDraw")} variant={data.drawVariant}>
                    <Icon name="PenLine" />
                    <Text>Draw</Text>
                  </Button>
                  <Button onClick={() => onAction("undo")} variant="icon" aria-label="Undo"><Icon name="Undo2" /></Button>
                  <Button onClick={() => onAction("redo")} variant="icon" aria-label="Redo"><Icon name="Redo2" /></Button>
                </Row>
                {data.cropTool && (
                  <>
                    <Grid columns={3} variant="controls">
                      <Field label="Aspect ratio"><Select name="aspect" options={data.aspectOptions} /></Field>
                      <Field label="Crop zoom"><Input name="cropZoom" type="range" min={1} max={3} step={0.01} /></Field>
                      <Field label="Straighten"><Input name="straighten" type="range" min={-15} max={15} step={0.5} /></Field>
                    </Grid>
                    <Row>
                      <Button onClick={() => onAction("rotate")} variant="secondary">Rotate</Button>
                      <Button onClick={() => onAction("flip")} variant="secondary">Flip</Button>
                    </Row>
                  </>
                )}
                {data.filterTool && (
                  <Grid columns={3} variant="controls">
                    <Field label="Filter"><Select name="filter" options={data.filterOptions} /></Field>
                    <Field label="Brightness"><Input name="brightness" type="range" min={-0.5} max={0.5} step={0.01} /></Field>
                    <Field label="Contrast"><Input name="contrast" type="range" min={-0.5} max={0.5} step={0.01} /></Field>
                    <Field label="Saturation"><Input name="saturation" type="range" min={-1} max={1} step={0.01} /></Field>
                    <Field label="Warmth"><Input name="warmth" type="range" min={-0.5} max={0.5} step={0.01} /></Field>
                    <Field label="Blur"><Input name="blur" type="range" min={0} max={0.3} step={0.01} /></Field>
                  </Grid>
                )}
                {data.textTool && (
                  <>
                    <Field label="Overlay text / emoji"><Input name="overlayText" multiline={true} placeholder="Write something… or add ✨" /></Field>
                    <Grid columns={3} variant="controls">
                      <Field label="Text size"><Input name="fontSize" type="number" min={12} max={160} /></Field>
                      <Field label="Font"><Select name="font" options={data.fontOptions} /></Field>
                      <Field label="Text / brush color"><Input name="color" type="color" /></Field>
                      <Checkbox name="textBackground">Text highlight</Checkbox>
                    </Grid>
                    <Row variant="edit-tools">
                      <Button onClick={() => onAction("addText")} variant="secondary">Add text</Button>
                      <Button onClick={() => onAction("addSticker")} variant="secondary">Add sticker</Button>
                      <Button onClick={() => onAction("updateText")} variant="secondary">Apply text style</Button>
                      <Button onClick={() => onAction("template")} variant="secondary">Story template</Button>
                    </Row>
                  </>
                )}
                {data.drawTool && (
                  <Row>
                    <Field label="Text / brush color"><Input name="color" type="color" /></Field>
                    <Field label="Brush width"><Input name="brushWidth" type="range" min={1} max={40} step={1} /></Field>
                    <Button onClick={() => onAction("brush")} variant="secondary">Draw / select</Button>
                  </Row>
                )}
                <Row variant="edit-tools">
                  <Button onClick={() => onAction("deleteLayer")} variant="text">Delete layer</Button>
                  <Button onClick={() => onAction("layerUp")} variant="text">Bring forward</Button>
                </Row>
                <Text variant="caption">Drag overlays to move them. Use the handles to resize and rotate.</Text>
              </>
            )}
          </Stack>
        )}
        {data.sharing && (
          <Stack variant="compose-share">
            <Row variant="caption-entry">
              <Section id="share-preview" />
              <Field label="Caption"><Input name="caption" multiline={true} placeholder="Add a caption…" /></Field>
            </Row>
            <Field label="Who can see this?"><Select name="visibility" options={data.audienceOptions} /></Field>
            <Field label="Location (optional)"><Input name="location" placeholder="Add location" /></Field>
            <Button onClick={() => onAction("togglePostOptions")} variant="option-row">
              {data.optionsLabel}
              <Icon name="ChevronDown" />
            </Button>
            {data.moreOptions && (
              <Stack variant="post-options">
                <Field label="Alt text"><Input name="alt" placeholder="Describe the image for screen readers" /></Field>
                <Field label="Schedule (optional)"><Input name="schedule" type="datetime-local" /></Field>
              </Stack>
            )}
            {data.isStory && (
              <>
                <Field label="Poll question (optional)"><Input name="poll" placeholder="What do you think?" /></Field>
                <Field label="Poll choices (comma separated)"><Input name="pollChoices" placeholder="Love it, Tell me more" /></Field>
                <Field label="Link sticker (HTTPS, optional)"><Input name="storyLink" type="url" /></Field>
              </>
            )}
            <Stack variant="media-preparation">
              <Text variant="caption">{data.prepareText}</Text>
              {data.preparing && (
                <Spinner variant="import-progress" value={data.preparePercent} label="Media upload progress" />
              )}
              {data.prepareError && (
                <>
                  <Alert tone="error">{data.prepareError}</Alert>
                  <Button onClick={() => onAction("retryPreparation")} variant="secondary">Retry upload</Button>
                </>
              )}
            </Stack>
            <Row variant="share-actions">
              <Button onClick={() => onAction("preview")} variant="secondary">Preview</Button>
              <Button onClick={() => onAction("publish")} variant="share">Share</Button>
            </Row>
          </Stack>
        )}
      </Stack>
    </Page>
  )
}
