# Project Name Brainstorm

**Date:** 2026-02-19
**Status:** Draft

---

## What We're Building

Renaming the CLI markdown reader from the placeholder name `cli-md` to **incipit**.

---

## Chosen Name

### incipit

Latin: *it begins* (third-person singular present of *incipere*).

The word used to open medieval manuscripts — the first words of a document, written in red ink (rubric) or ornamental lettering to mark the start of reading. Functionally: every time you run `incipit README.md`, you are beginning to read a document. The name is the action.

```
$ incipit README.md
```

---

## Why This Name

- **Unique:** No software or software company uses this name (as of 2026-02-19)
- **Literary, not ebook-adjacent:** Not associated with the ebook/library app space (unlike tome, folio, codex, vellum, scroll)
- **Self-describing without being literal:** Evokes "beginning to read" without spelling out "markdown reader"
- **Pronounceable and typeable:** in-SIP-it or IN-si-pit — two syllables, no ambiguity
- **Works as a command:** `incipit README.md` reads naturally as a sentence

---

## Rejected Alternatives

| Name | Reason rejected |
|---|---|
| tome | Taken by ebook/library software |
| folio | Taken by publishing/library software |
| codex | Associated with GitHub Copilot (OpenAI Codex) |
| vellum | Taken by ebook/writing software |
| scroll | Taken by productivity/ebook software |
| verso | Strong contender; rejected because incipit has better semantic fit |
| fleuron | Beautiful but harder to pronounce/type |
| hedera | Obscure; "hedera" is also a blockchain company |
| rune | Short but too associated with ancient Norse mysticism / gaming |
| nib | Too minimal; no connection to reading |

---

## Key Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Name | **incipit** | Latin for "here begins"; unique, literary, self-describing |
| Binary name | `incipit` | Matches project name; natural CLI command |
| Repository name | `incipit` | Replace `cli-md` |

---

## Open Questions

None.

---

## Resolved Questions

- **Vibe?** → Literary / book-ish
- **Abstract or hint?** → Subtle hint (evokes reading without being literal)
- **Uniqueness constraint?** → Must not exist as any software or software company name
