# Task Schema

Each task is a markdown file with YAML frontmatter.

## Required fields
- `type: task`
- `title`: short summary
- `status`: one of backlog, todo, in_progress, review, blocked, done, cancelled
- `priority`: integer 0-4 (0 = urgent, 4 = low)

## Optional fields
- `assignee`: agent or human identifier
- `blocked-by`: list of task paths that must be `done` before this task starts
- `due`: ISO date
- `tags`: list of strings
- `claimed-by`: set automatically by claim endpoint
- `claimed-at`: set automatically by claim endpoint
- `lease-expires`: set automatically by claim endpoint
