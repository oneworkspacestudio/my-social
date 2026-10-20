// Generated from screens/Camera.aui; edit the AUI source.
import { Badge, Button, Card, Checkbox, Field, Input, Page, Row, Section, Select, Spacer, Text } from "@/components/ui"

export function Camera({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Card variant="camera-panel">
        <Row>
          <Badge tone="warning">{data.cameraStatus}</Badge>
          <Spacer />
          <Button onClick={() => onAction("closeCamera")} variant="secondary">Close camera</Button>
        </Row>
        <Section id="camera-video" />
        <Row>
          <Button onClick={() => onAction("switchCamera")} variant="secondary">Switch camera</Button>
          <Button onClick={() => onAction("shutter")}>Take photo</Button>
          <Button onClick={() => onAction("record")} variant="danger">{data.recordLabel}</Button>
          {data.canPause && (
            <Button onClick={() => onAction("pauseRecord")} variant="secondary">Pause / resume recording</Button>
          )}
        </Row>
        <Row>
          <Checkbox name="mic">Enable microphone before recording</Checkbox>
          <Field label="Aspect guide"><Select name="aspect" options={data.aspectOptions} /></Field>
        </Row>
        {data.hasZoom && (
          <Field label="Camera zoom"><Input name="cameraZoom" type="range" min={data.zoomMin} max={data.zoomMax} step={data.zoomStep} /></Field>
        )}
        {data.hasTorch && (
          <Button onClick={() => onAction("torch")} variant="secondary">Toggle device torch</Button>
        )}
        <Text variant="caption">The camera stops when you leave this screen or put the app in the background. Audio is captured only while Enable microphone is checked.</Text>
      </Card>
    </Page>
  )
}
