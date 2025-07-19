-- This table will store permissions for document sharing.
CREATE TABLE document_permissions (
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- 'role' can be used for future enhancements (e.g., 'viewer', 'editor')
    -- For now, any entry means they have access.
    role VARCHAR(20) NOT NULL DEFAULT 'editor',
    
    -- A user can only have one role per document.
    PRIMARY KEY (document_id, user_id)
);

-- Create an index for efficient lookups by user.
CREATE INDEX idx_document_permissions_user_id ON document_permissions(user_id);