# Yafa CLI help
command({command: "..."}). Up to 4 commands, joined by &&; stop on failure.
! required; ? optional (default null); = default. Quote text with spaces.
--flag value or --flag=value. Bare BOOL flag means true. null keeps/omits nullable fields; "null" is literal text. Backslash escapes in either quote style: \n \r \t \\ \" \'. Literal newlines are not allowed in the command envelope.
No shell, pipes, redirects, substitutions or variable expansion. Quoted && is text.
Use actual returned IDs/revisions. Chain only commands whose arguments are already known; dependent commands need a later tool call.
Syntax/types/enums/ranges are checked for the whole chain first. Domain rules run in order; earlier writes remain committed after a later failure. Read actual state after conflict; never blindly replay writes. Each command consumes one tool slot; a rejected envelope consumes one slot.

{{command_reference}}

Dates: agenda.day/state.read: YYYY-MM-DD|today|yesterday (agenda.day also tomorrow). agenda.week --week-of: YYYY-MM-DD|this_week|last_week|next_week. agenda.month --month: YYYY-MM|this_month|last_month|next_month. Omitted selector = current Paris day/week/month. agenda.range endpoints inclusive, at most 31 days. Page with returned next_offset.
Me = human tasks; Yafa = indexed thread activity. Undated tasks: tasks.list. No external calendar or complete-history guarantee.
Task --scheduled-date: YYYY-MM-DD; --scheduled-time: HH:mm, requires day. --due-at: separate RFC3339 deadline, never infer one. Creating tasks does not schedule notifications.
context names: soul.md|directive.md|project.md|workspace.md|user.md. Read before editing; edits apply next run/session.
state.record: subjective estimates only, no future date. Quote current user words (max 300 bytes; Live: start of [user] segment). Null fields keep observations; initial revision is "0". Do not infer mood from task completion.

{{optional_commands_help}}
