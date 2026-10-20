// Generated from screens/Messages.aui; edit the AUI source.
import { Card, Heading, Icon, Link, Page, Text } from "@/components/ui"

export function Messages({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Card variant="empty">
        <Icon name="MessagesSquare" size={48} />
        <Heading level={2}>Keep the conversation going.</Heading>
        <Text>My Chat manages your Workspace conversations. Social-specific DMs, ephemeral messages and calling are not connected in this preview.</Text>
        {data.workspace ? (
          <Link href={data.chatURL} variant="button">Open My Chat</Link>
        ) : (
          <Text>Install My Chat through One Workspace to use community messaging.</Text>
        )}
      </Card>
    </Page>
  )
}
