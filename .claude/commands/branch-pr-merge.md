Create a branch, commit, PR, and merge for the given issue.

## Instructions

Given an issue number and the staged/unstaged changes in the working tree:

1. **Determine branch name** from the issue:
   - Read the issue title: `gh issue view <number> --json title -q '.title'`
   - Use `fix/<number>-<short-slug>` for bugs, `feat/<number>-<short-slug>` for features, `test/<number>-<short-slug>` for test-only changes

2. **Create branch from main**:
   ```bash
   git checkout main && git pull origin main
   git checkout -b <branch-name> main
   ```

3. **Stage and commit** the relevant changes:
   - Only stage files related to this specific issue
   - Use conventional commit format: `fix:`, `feat:`, `test:`
   - Include `Closes #<number>` in the commit body
   - End with `Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>`
   - Use HEREDOC for multi-line commit messages

4. **Push and create PR**:
   ```bash
   git push -u origin <branch-name>
   gh pr create --base main --head <branch-name> --title "<commit title>" --body "<summary with Closes #N>"
   ```

5. **Merge and clean up**:
   ```bash
   gh pr merge <pr-number> --merge --delete-branch
   git checkout main && git pull origin main
   ```

6. **Verify**: Run `go test ./...` before committing if Go files changed.

## Rules

- If multiple issues touch the same files and are tightly coupled, combine them in one branch with multiple `Closes #N`
- Never force-push
- Always delete branch after merge
- Always return to main after merge
- PR body must include `## Summary` with bullet points

## Input

$ARGUMENTS
