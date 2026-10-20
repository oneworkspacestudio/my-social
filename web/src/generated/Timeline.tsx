// Generated from screens/Timeline.aui; edit the AUI source.
import { Button, Checkbox, Field, Heading, Input, Page, Row, Select, Stack, Text } from "@/components/ui"

export function Timeline({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Stack>
        <Heading level={3}>Your Reel timeline</Heading>
        <Text>Trim, reorder and adjust clip speed. Renders use 720 × 1280 H.264, up to 90 seconds. Multi-clip renders currently have no audio.</Text>
        {data.clips.map((item, i) => (
          <Row key={i} variant="clip-row">
            <Text>{item.title}</Text>
            <Text variant="caption">{item.trim}</Text>
            <Button onClick={() => onAction("selectClip")} item={item.id} variant="secondary">Edit clip</Button>
            <Button onClick={() => onAction("clipEarlier")} item={item.id} variant="secondary">Move earlier</Button>
            <Button onClick={() => onAction("clipRemove")} item={item.id} variant="secondary">Remove</Button>
          </Row>
        ))}
        <Row>
          <Field label="Trim start (seconds)"><Input name="clipStart" type="number" min={0} step={0.1} /></Field>
          <Field label="Trim end (seconds)"><Input name="clipEnd" type="number" min={0.1} step={0.1} /></Field>
          <Field label="Playback speed"><Select name="clipSpeed" options={data.speedOptions} /></Field>
        </Row>
        <Row>
          <Button onClick={() => onAction("clipApply")} variant="secondary">Apply trim</Button>
          <Checkbox name="muteVideo">Mute original audio</Checkbox>
          <Field label="Video quality"><Select name="videoQuality" options={data.qualityOptions} /></Field>
          <Button onClick={() => onAction("renderVideo")}>Render Reel</Button>
        </Row>
        <Field label="Captions (WebVTT, optional)"><Input name="captions" multiline={true} placeholder="WEBVTT" /></Field>
        <Text variant="caption">{data.renderStatus}</Text>
      </Stack>
    </Page>
  )
}
