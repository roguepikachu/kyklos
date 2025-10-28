# Documentation Style Guide

This guide defines the writing standards for Kyklos documentation.

## Voice and Tone

### Clear and Direct

Use simple, precise language. Avoid jargon unless defined.

**Good:**
> The controller scales your deployment based on time windows.

**Avoid:**
> The reconciliation loop leverages temporal heuristics to dynamically orchestrate workload capacity.

### Active Voice

Write with active verbs. Tell the reader what to do.

**Good:**
> Set the timezone to match your business location.

**Avoid:**
> The timezone should be set to match your business location.

### Confident

State facts directly without hedging.

**Good:**
> Kyklos handles DST transitions automatically.

**Avoid:**
> Kyklos should handle DST transitions in most cases, typically.

### Helpful

Anticipate confusion and address it proactively.

**Good:**
> The end time is exclusive. At 17:00:00 exactly, the window has already ended.

**Better:**
> The end time is exclusive. At 17:00:00 exactly, the window has already ended. This prevents gaps when windows connect: one ends at 17:00, the next starts at 17:00.

## Sentence Structure

### Short Sentences

Aim for 15-20 words per sentence. Break long sentences into two.

**Good:**
> Windows define when to scale. Each window specifies days, times, and replica count.

**Avoid:**
> Windows, which define when to scale, each specify days of the week when they apply, the start and end times for the window, and the desired replica count during that time period.

### One Idea Per Sentence

Each sentence should convey a single concept.

**Good:**
> Grace periods delay downscaling. They do not affect scale-ups.

**Avoid:**
> Grace periods delay downscaling but not scale-ups.

### Consistent Tense

Use present tense for current behavior. Use future tense for planned features.

**Good:**
> The controller reconciles every 30 seconds. Future versions will support webhooks.

## Section Organization

### Short Sections

Keep sections to 3-5 paragraphs maximum. Break long content into subsections.

### Descriptive Headings

Headings should answer a question or name a specific topic.

**Good:**
> ### How Do Cross-Midnight Windows Work?

**Avoid:**
> ### Windows

### Progressive Disclosure

Start simple, add complexity gradually.

1. State what something is
2. Show a basic example
3. Explain edge cases
4. Link to detailed docs

**Example structure:**
```markdown
## Time Windows

A time window defines when your application needs more capacity.

### Basic Example
### Cross-Midnight Windows
### Overlapping Windows
### See Also: [API Reference]
```

## Examples

### Examples First, Theory Second

Show concrete examples before abstract explanations.

**Good:**
```markdown
## Grace Period

Grace periods delay downscaling:

```yaml
gracePeriodSeconds: 300  # 5 minutes
```

When leaving a window, the controller waits 5 minutes before reducing replicas.
```

**Avoid:**
```markdown
## Grace Period

A grace period is a temporal buffer mechanism that defers the execution of scale-down operations...
```

### Complete, Working Examples

Every code example should run without modification.

**Good:**
```yaml
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: office-hours
  namespace: production
spec:
  targetRef:
    kind: Deployment
    name: webapp
  timezone: America/New_York
  defaultReplicas: 2
  windows:
  - days: [Mon, Tue, Wed, Thu, Fri]
    start: "09:00"
    end: "17:00"
    replicas: 10
```

**Avoid:**
```yaml
spec:
  windows:
  - days: [...]
    start: "..."
    # add your times here
```

### Annotate Examples

Add comments explaining non-obvious parts.

```yaml
windows:
- days: [Fri]
  start: "22:00"  # Friday night
  end: "06:00"    # Saturday morning
  replicas: 5
```

## Commands and Code

### Commands Before Explanation

Show the command, then explain what it does.

**Good:**
```bash
kubectl get tws -n production

# This lists all TimeWindowScalers in the production namespace
```

**Avoid:**
> To list TimeWindowScalers, you would typically use the kubectl command with the get verb, specifying the resource type and namespace.

### Expected Output

Show what success looks like.

```bash
kubectl get pods -n kyklos-system
```

Expected output:
```
NAME                                      READY   STATUS    RESTARTS   AGE
kyklos-controller-manager-abc123-xyz      1/1     Running   0          30s
```

### Full Command Paths

Use absolute paths for commands that might be ambiguous.

**Good:**
```bash
make cluster-up
```

**Good for emphasis:**
```bash
cd /Users/aykumar/personal/kyklos
make cluster-up
```

## Links and References

### Link Text

Use descriptive link text, not "click here."

**Good:**
> See the [Operations Guide](OPERATIONS.md) for production metrics.

**Avoid:**
> For production metrics, [click here](OPERATIONS.md).

### Internal Links

Link to related concepts generously.

**At first mention:**
> The effective replicas (see [Effective Replicas](#effective-replicas)) determine...

**In "See Also" sections:**
> - [Concepts](CONCEPTS.md) - Window matching details
> - [FAQ](FAQ.md) - Common questions

### External Links

Minimize external links. Capture essential info in docs.

**Acceptable:**
> For IANA timezone identifiers, see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones

**Better:**
> Use IANA timezone identifiers like `America/New_York`, `Europe/London`, or `Asia/Kolkata`.

## Formatting

### Bold for Emphasis

Use bold sparingly for key concepts.

**Good:**
> The **last matching window wins** when multiple windows overlap.

### Code Formatting

Use backticks for:
- Field names: `spec.timezone`
- Commands: `kubectl get tws`
- File names: `CONCEPTS.md`
- Values: `defaultReplicas: 2`

### Lists

Use bulleted lists for unordered items:
- First item
- Second item
- Third item

Use numbered lists for sequences:
1. Create cluster
2. Deploy controller
3. Apply TimeWindowScaler

### Tables

Use tables for structured comparisons.

| Window | Replicas | Days |
|--------|----------|------|
| Morning | 5 | Mon-Fri |
| Peak | 10 | Mon-Fri |
| Evening | 3 | Mon-Fri |

## Term Usage

### Consistent Terminology

Use the same term for the same concept throughout.

**Correct:**
- TimeWindowScaler (not "scaler", "TWS resource", "time window object")
- Deployment (not "deploy", "workload", "target")
- Effective replicas (not "desired replicas", "wanted replicas")

### Define Before Using

Introduce terms before using them extensively.

**Good:**
> Kyklos creates a TimeWindowScaler (TWS) for each scaled deployment. Each TWS defines...

### Use Glossary

For definitions, link to the glossary.

> See the [Glossary](GLOSSARY.md) for term definitions.

## Special Content

### Warnings

Use bold "Warning:" for destructive operations.

**Warning:** This command deletes all TimeWindowScalers. This cannot be undone.

### Notes

Use plain "Note:" for additional context.

**Note:** Cross-midnight windows use the starting day only.

### Examples

Label examples clearly.

**Example:** Friday night to Saturday morning window

```yaml
- days: [Fri]
  start: "22:00"
  end: "06:00"
  replicas: 5
```

## Document Structure

### Standard Sections

Most docs should include:

1. **Title** - Clear, descriptive
2. **Introduction** - One paragraph overview
3. **Main Content** - Progressive complexity
4. **Examples** - Concrete demonstrations
5. **See Also** - Links to related docs

### README Structure

```markdown
# Title

[One paragraph description]

## Why [Project]?

[Benefits, bullet points]

## Quick Start

[Under 5 minutes, commands only]

### Prerequisites
### Installation
### Your First [Resource]

## Next Steps

[Links to docs]

## Project Status

[Version, features, limitations]
```

### Concept Doc Structure

```markdown
# [Concept Name]

[One paragraph introduction]

## Basic Concept

[Simplest explanation + example]

## Advanced Usage

[Edge cases, complex scenarios]

## Best Practices

[Recommendations]

## See Also

[Links]
```

### How-To Guide Structure

```markdown
# How to [Task]

[When and why you'd do this]

## Steps

1. [First step with command]
2. [Second step with command]
3. [Verification with expected output]

## Troubleshooting

[Common issues]
```

## Writing Process

### Write for Scanning

Readers scan before reading deeply.

**Support scanning with:**
- Clear headings
- Short paragraphs
- Bulleted lists
- Bold key terms
- Code examples

### Review Checklist

Before publishing, verify:

- [ ] Commands are copy-pasteable
- [ ] Examples run without modification
- [ ] Terms are consistent
- [ ] Links work
- [ ] No jargon without definitions
- [ ] Sentences under 20 words
- [ ] One idea per sentence
- [ ] Active voice
- [ ] Present tense

## Anti-Patterns

### Avoid These

**Hedging:**
> Kyklos should generally handle most cases.

**Correct:**
> Kyklos handles DST transitions automatically.

---

**Passive voice:**
> The deployment is scaled by the controller.

**Correct:**
> The controller scales the deployment.

---

**Jargon:**
> Leverage the reconciliation loop to actuate capacity transformations.

**Correct:**
> The controller reconciles to change replica counts.

---

**Vague examples:**
```yaml
replicas: X  # set this value
```

**Correct:**
```yaml
replicas: 10  # Peak business hours capacity
```

---

**Long sentences:**
> The grace period, which is an optional feature that you can configure via the gracePeriodSeconds field, delays scaling operations when the replica count would decrease, but only for downscaling operations and not for scale-ups which happen immediately.

**Correct:**
> Grace periods delay downscaling only. Configure via `gracePeriodSeconds`. Scale-ups happen immediately.

---

**Hidden commands:**
> You can list TimeWindowScalers by invoking the Kubernetes API via kubectl with the get verb.

**Correct:**
```bash
kubectl get tws --all-namespaces
```

## Document Types

### README
- Hook readers in one paragraph
- Get them running in 5 minutes
- Link to deeper docs

### Concepts
- Explain how things work
- Use analogies and examples
- Link to API reference for details

### How-To Guides
- Numbered steps
- Commands first, explanation second
- Expected output after each step

### Reference
- Complete, accurate
- Organized for lookup (alphabetical, grouped)
- Tables for structured data

### Troubleshooting
- Symptom-based organization
- Diagnosis steps
- Clear resolution procedures

## Final Principles

1. **Examples over prose** - Show, don't just tell
2. **Short everything** - Sentences, paragraphs, sections
3. **Active voice** - Tell the reader what to do
4. **No jargon** - Or define it on first use
5. **Working examples** - Every example must run
6. **Consistent terms** - Same word for same concept
7. **Progressive disclosure** - Simple first, complex later
8. **Scannable structure** - Headings, lists, bold
9. **Helpful tone** - Anticipate confusion
10. **Accuracy over completeness** - Better to document less correctly

## See Also

- [Glossary](GLOSSARY.md) - Term definitions for consistent usage
- [API Reference](../api/CRD-SPEC.md) - Technical accuracy reference
- [Concepts](CONCEPTS.md) - Example of progressive disclosure
