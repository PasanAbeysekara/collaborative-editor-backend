---
name: Feature Request
about: Suggest an idea for this project
title: 'Feat: [BRIEF FEATURE DESCRIPTION]'
labels: 'enhancement, feature'
assignees: ''

---

**Is your feature request related to a problem? Please describe.**
A clear and concise description of what the problem is. *(e.g., "After I undo a change, I sometimes realize I didn't want to, and there is no way to re-apply the change.")*

**Describe the solution you'd like**
A clear and concise description of what you want to happen.
*(e.g., "I would like a new `redo` operation type. When the server receives this operation, it should re-apply the last operation that was undone.")*

**Describe alternatives you've considered**
A clear and concise description of any alternative solutions or features you've considered.

**Acceptance Criteria**
- [ ] A new `OpRedo` type is added.
- [ ] A new Redis list (e.g., `doc_redo_ops:{docID}`) is used to store undone operations.
- [ ] When a user sends an `undo` op, the undone op is moved from the `ops` list to the `redo_ops` list.
- [ ] When a user sends a `redo` op, the last op is popped from the `redo_ops` list, applied, and broadcast.
- [ ] After a new `insert` or `delete` operation is applied, the `redo_ops` list for that document should be cleared.

**Additional context**
Add any other context or mockups about the feature request here.