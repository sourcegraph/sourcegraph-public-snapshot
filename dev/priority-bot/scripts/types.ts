export interface slackMessage {
  blocks?: (BlocksEntity)[] | null;
}
interface BlocksEntity {
  type: string;
  text?: TextOrFieldsEntity | null;
  fields?: (TextOrFieldsEntity1)[] | null;
}
interface TextOrFieldsEntity {
  type: string;
  text: string;
}
interface TextOrFieldsEntity1 {
  type: string;
  text: string;
}


export interface GithubIssue {
    active_lock_reason?: null;
    assignee?: null;
    assignees?: (null)[] | null;
    author_association: string;
    body: string;
    closed_at?: null;
    comments: number;
    comments_url: string;
    created_at: string;
    events_url: string;
    html_url: string;
    id: number;
    labels?: (LabelsEntity)[] | null;
    labels_url: string;
    locked: boolean;
    milestone?: null;
    node_id: string;
    number: number;
    performed_via_github_app?: null;
    reactions: Reactions;
    repository_url: string;
    state: string;
    timeline_url: string;
    title: string;
    updated_at: string;
    url: string;
    user: User;
  }
  interface LabelsEntity {
    color: string;
    default: boolean;
    description?: string | null;
    id: number;
    name: string;
    node_id: string;
    url: string;
  }
  interface Reactions {
    confused: number;
    eyes: number;
    heart: number;
    hooray: number;
    laugh: number;
    rocket: number;
    total_count: number;
    url: string;
  }
  interface User {
    avatar_url: string;
    events_url: string;
    followers_url: string;
    following_url: string;
    gists_url: string;
    gravatar_id: string;
    html_url: string;
    id: number;
    login: string;
    node_id: string;
    organizations_url: string;
    received_events_url: string;
    repos_url: string;
    site_admin: boolean;
    starred_url: string;
    subscriptions_url: string;
    type: string;
    url: string;
  }

export interface SlackJsonTemplate {
    needsPriority: boolean;
    needsEstimate: boolean;
}