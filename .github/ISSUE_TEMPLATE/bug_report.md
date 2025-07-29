---
name: Bug Report
about: Create a report to help us improve the application
title: 'Bug: [BRIEF DESCRIPTION OF THE BUG]'
labels: 'bug, triage'
assignees: ''

---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior. This is the most important section!
1. Login as User A and get a token (`TOKEN_A`).
2. Login as User B and get a token (`TOKEN_B`).
3. Create a document as User A, get its ID (`DOCUMENT_ID`).
4. Share the document with User B.
5. Connect to the WebSocket as User A (`wscat ...`).
6. Send the following operation: `...`
7. Connect to the WebSocket as User B.
8. Send the following operation: `...`

**Expected behavior**
A clear and concise description of what you expected to happen.
*(e.g., "I expected the document content to revert to 'Hello' and the server to broadcast a `delete` operation.")*

**Actual behavior**
A clear and concise description of what actually happened.
*(e.g., "The undo operation was ignored. No message was broadcast from the server, and the state in Redis and PostgreSQL remained 'Hello World'.")*

**Screenshots or Logs**
If applicable, add screenshots, terminal output, or server logs to help explain your problem.