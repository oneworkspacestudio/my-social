// Generated from screens/Feed.aui; edit the AUI source.
import { Badge, Button, Card, Heading, Icon, Image, Input, Link, Page, Row, Section, Spacer, Stack, Text } from "@/components/ui"

export function Feed({ data, onAction }: { data: any; onAction: (name: string) => void }) {
  return (
    <Page>
      <Stack variant="feed-layout">
        <Stack variant="feed-column">
          {data.isHome && (
            <>
              <Card variant="story-strip">
                <Button variant="story-add" onClick={() => onAction("createStory")} aria-label="Create a Story">
                  <Icon name="Plus" />
                  <Text>Your Story</Text>
                </Button>
                {data.stories.map((item, i) => (
                  <Button key={i} variant="story" item={item.id} onClick={() => onAction("viewStory")} aria-label={item.profile.handle}>
                    <Image src={item.thumbnail} alt="" width={56} height={56} />
                    <Text>{item.profile.handle}</Text>
                  </Button>
                ))}
              </Card>
              <Row variant="feed-tabs">
                <Link href="#home" variant={data.homeVariant}>For you</Link>
                <Link href="#following" variant={data.followingVariant}>Following</Link>
                <Spacer />
                <Text variant="caption">A little closer, every day.</Text>
              </Row>
            </>
          )}
          {data.isExplore && (
            <>
              <Row>
                <Input name="query" type="search" placeholder="Search people, captions and #hashtags" />
                <Button onClick={() => onAction("search")}>Search</Button>
              </Row>
              <Row variant="people">
                {data.people.map((item, i) => (
                  <Link key={i} href={item.href} variant="person">
                    {item.avatarURL && (
                      <Image src={item.avatarURL} variant="profile-photo" alt="Profile photo" width={40} height={40} />
                    )}
                    <Text>{item.handle}</Text>
                    <Text variant="caption">{item.bio}</Text>
                  </Link>
                ))}
              </Row>
            </>
          )}
          {data.isProfile && (
            <Card variant="profile-header">
              {data.profile.avatarURL ? (
                <Image src={data.profile.avatarURL} variant="profile-photo" alt="Profile photo" width={96} height={96} />
              ) : (
                <Text variant="profile-avatar">{data.profile.initials}</Text>
              )}
              <Text variant="eyebrow">{data.profile.handle}</Text>
              <Heading level={2}>{data.profile.name}</Heading>
              <Text>{data.profile.bio}</Text>
              <Text variant="caption">{data.profile.counts}</Text>
              {data.otherProfile && (
                <>
                  <Button onClick={() => onAction("follow")}>{data.followLabel}</Button>
                  <Button variant="secondary" onClick={() => onAction("block")}>Block / unblock</Button>
                  <Button variant="secondary" onClick={() => onAction("mute")}>Mute / unmute</Button>
                  <Button variant="secondary" onClick={() => onAction("closeFriend")}>Close Friends / remove</Button>
                  <Button variant="secondary" onClick={() => onAction("restrict")}>Restrict / unrestrict</Button>
                </>
              )}
              {data.otherProfile ? (
                <Text variant="caption" />
              ) : (
                <Link href="#settings" variant="edit-profile">Edit profile & photo</Link>
              )}
              {data.highlights.map((item, i) => (
                <Button key={i} onClick={() => onAction("highlight")} item={item.id} variant="secondary">{item.title}</Button>
              ))}
            </Card>
          )}
          <Stack variant={data.feedStyle}>
            {data.posts.map((item, i) => (
              <Card key={i} variant="post">
                <Row variant="post-header">
                  <Link href={item.profile.href} variant="avatar">
                    {item.profile.avatarURL ? (
                      <Image src={item.profile.avatarURL} variant="avatar" alt="Profile photo" width={44} height={44} />
                    ) : (
                      <Text>{item.profile.initials}</Text>
                    )}
                  </Link>
                  <Stack gap="xs">
                    <Link href={item.profile.href} variant="handle">{item.profile.handle}</Link>
                    <Text variant="caption">{item.date}</Text>
                  </Stack>
                  <Spacer />
                  <Badge tone="muted">{item.visibility}</Badge>
                </Row>
                <Section id="post-media" item={item} />
                <Row variant="post-actions">
                  <Button item={item.id} onClick={() => onAction("like")} variant={item.likeVariant} aria-label="Like post"><Icon name="Heart" /></Button>
                  <Button item={item.id} onClick={() => onAction("comments")} variant="icon" aria-label="View comments"><Icon name="MessageCircle" /></Button>
                  <Button item={item.id} onClick={() => onAction("share")} variant="icon" aria-label="Share post"><Icon name="Send" /></Button>
                  <Spacer />
                  <Button item={item.id} onClick={() => onAction("save")} variant={item.saveVariant} aria-label="Save post"><Icon name="Bookmark" /></Button>
                </Row>
                <Text variant="handle">{item.likeText}</Text>
                <Text variant="post-caption">{item.caption}</Text>
                <Button item={item.id} onClick={() => onAction("comments")} variant="text">{item.commentText}</Button>
                {item.own ? (
                  <Row>
                    <Button item={item.id} onClick={() => onAction("archive")} variant="text">Archive</Button>
                    <Button item={item.id} onClick={() => onAction("editPost")} variant="text">Edit details</Button>
                  </Row>
                ) : (
                  <Button item={item.id} onClick={() => onAction("report")} variant="text">Report</Button>
                )}
              </Card>
            ))}
          </Stack>
          {data.empty && (
            <Card variant="empty">
              <Icon name="Camera" size={40} />
              <Heading level={2}>{data.emptyTitle}</Heading>
              <Text>{data.emptyText}</Text>
              <Button onClick={() => onAction("create")}>Create your first moment</Button>
            </Card>
          )}
          {data.hasMore && (
            <Button onClick={() => onAction("more")} variant="secondary">Load more</Button>
          )}
        </Stack>
        {data.isHome && (
          <Stack variant="right-rail">
            <Card variant="welcome">
              <Badge tone="success">YOUR OWN LITTLE CORNER</Badge>
              <Heading level={2}>Good moments deserve good company.</Heading>
              <Text>Share something you made, somewhere you went, or a small part of your day.</Text>
              <Button onClick={() => onAction("create")}>Make a moment</Button>
            </Card>
            <Card>
              <Heading level={3}>People to know</Heading>
              {data.suggested.map((item, i) => (
                <Row key={i} variant="suggestion">
                  <Link href={item.href} variant="avatar">
                    {item.avatarURL ? (
                      <Image src={item.avatarURL} variant="avatar" alt="Profile photo" width={44} height={44} />
                    ) : (
                      <Text>{item.initials}</Text>
                    )}
                  </Link>
                  <Stack gap="xs">
                    <Link href={item.href} variant="handle">{item.handle}</Link>
                    <Text variant="caption">{item.name}</Text>
                  </Stack>
                </Row>
              ))}
            </Card>
            <Text variant="caption">Made for your community. Hosted by you.</Text>
          </Stack>
        )}
      </Stack>
    </Page>
  )
}
