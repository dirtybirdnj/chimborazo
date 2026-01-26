# Integration Ceremony - Chimborazo

**Status:** Pending (scheduled for Monday)
**Last Updated:** 2024-12-20

---

## The Philosophy

> "Breaking the build is not a technical failure - it's a process failure. The code didn't betray you. The ceremony was skipped."

### Old Way vs New Way

| Old Way (Cowboy Deployment) | New Way (Integration Ceremony) |
|----------------------------|-------------------------------|
| "Works on my machine" | "Works on the integrated branch" |
| Merge and pray | Rebase, test, validate, then merge |
| 3am pager alerts | Boring, predictable deploys |
| Hope-driven development | Ceremony-driven development |
| Individual heroics | Team rituals |

---

## Branch Inventory (as of 2024-12-20)

### Main (Production Baseline)
```
9398981 Initial scaffold: Cartography at the speed of code
```
Only the initial scaffold. All work is on feature branches.

### Feature Branches

| Branch | Commits Ahead | Status | Notes |
|--------|--------------|--------|-------|
| `feature/recipe-parser` | 4 | **PRIMARY** | Has all integrated work |
| `feature/svg-writer` | 1 | Superseded | 48 lines, basic skeleton |
| `feature/http-fetcher` | 1 | Superseded | Similar to recipe-parser version |
| `feature/cache-path-helper` | 1 | **ORPHAN?** | `cache.go` NOT in recipe-parser |

### Key Finding: Potential Orphaned Work

The `feature/cache-path-helper` branch has `internal/sources/cache.go` which does NOT exist in `feature/recipe-parser`. This needs investigation before merge.

---

## Integration Steps for Monday

### Step 1: Pre-Flight Check (clood tools)

```bash
# Orient yourself
clood preflight
clood tree ~/Code/chimborazo --depth 2

# Check branch states
cd ~/Code/chimborazo
git fetch --all
git branch -a
git log --oneline --all --graph -20
```

### Step 2: Investigate the Orphan

```bash
# What's in cache-path-helper that we might be missing?
git show feature/cache-path-helper:internal/sources/cache.go

# Is this functionality duplicated elsewhere in recipe-parser?
clood grep "CachePath" ~/Code/chimborazo
```

**Decision required:** Cherry-pick the orphan, or confirm it's superseded.

### Step 3: Create Integration Branch

```bash
# Never integrate directly on main
git checkout main
git pull origin main
git checkout -b integration/summit-prep

# Bring in the primary work
git merge feature/recipe-parser --no-ff
```

### Step 4: Test the Integration

```bash
# Build
go build ./...

# Run all tests
go test ./...

# Validate a recipe
./chimborazo validate recipes/test_vt.yaml

# Build a map
./chimborazo build recipes/test_vt.yaml

# Inspect output
open output/test_vt.svg
```

### Step 5: Validate Acceptance Criteria

Check against Epic #10 issues:

- [ ] #11 - census: source scheme (NOT YET)
- [ ] #12 - Subtract operation (NOT YET)
- [ ] #13 - Dynamic bounds (NOT YET)
- [ ] #15 - file: scheme (DONE)
- [ ] #16 - url: scheme (DONE)
- [ ] Recipe loads and validates (DONE)
- [ ] SVG renders correctly (DONE)

### Step 6: Planning Ceremony with User

Present findings:
1. What works
2. What's missing
3. What's the risk of merging
4. Recommendation

**Get explicit approval before proceeding.**

### Step 7: The Merge (with ceremony)

```bash
# Only after approval
git checkout main
git merge integration/summit-prep --no-ff -m "feat: Chimborazo summit prep - recipe parser, pipeline, SVG output

Integrates:
- Recipe parser with YAML loading and validation
- HTTP fetcher with caching (thoreau pattern)
- Geometry operations (clip, simplify, merge)
- SVG writer with proper projection
- Pipeline builder connecting all components

Closes #15, #16

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

# Push with confidence, not hope
git push origin main
```

### Step 8: Cleanup

```bash
# Archive merged branches
git branch -d feature/recipe-parser
git push origin --delete feature/recipe-parser

# Keep orphan branches until decision
# git branch -d feature/cache-path-helper  # ONLY if confirmed superseded
```

---

## Why This Matters

### Breaking the Build

In the old world, "breaking the build" was a dreaded event:
- CI/CD pipelines screamed red
- Slack channels exploded
- Someone got blamed

In the clood world, we prevent breaks through ceremony:
- Integration branches catch conflicts early
- Local testing before merge
- User approval as a gate

### Deploying to Production

"Production" for Chimborazo is `main` branch. Deploying means:
- Users pulling the repo get working code
- The README promises aren't lies
- `go build && ./chimborazo build recipes/example.yaml` actually works

The ceremony ensures every merge to main is a **promise kept**.

---

## Contrast: Old vs Clood Methods

### Old Method
1. Developer works in isolation
2. "git push origin main" (YOLO)
3. CI fails, or worse, passes but breaks in production
4. Scramble to fix
5. Blame, shame, 3am pages

### Clood Method
1. Work on feature branch
2. Use `clood grep/tree/symbols` to understand integration points
3. Create integration branch
4. Test locally with full ceremony
5. Use `clood ask` to review approach if uncertain
6. Present to user for approval
7. Merge with confidence
8. Boring, successful deploy

The goal is **boring deploys**. Excitement in production is bad.

---

## For Monday

1. Review this document
2. Run the pre-flight checks
3. Investigate the cache-path-helper orphan
4. Create integration branch
5. Run full test suite
6. Present findings for approval
7. Merge (or iterate)

The summit isn't going anywhere. Take the time to do it right.

---

*"Software integration is a dish best served with proper ceremony, not hope."*
