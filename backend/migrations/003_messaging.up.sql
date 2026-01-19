-- Conversations (two participants per conversation)
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    participant_one UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    participant_two UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_message_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(participant_one, participant_two),
    CHECK (participant_one < participant_two)  -- consistent ordering
);

-- Index for finding conversations by participant
CREATE INDEX idx_conversations_participant_one ON conversations(participant_one);
CREATE INDEX idx_conversations_participant_two ON conversations(participant_two);
CREATE INDEX idx_conversations_last_message ON conversations(last_message_at DESC);

-- Messages with encrypted content
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content_encrypted BYTEA NOT NULL,
    content_nonce BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for fetching messages by conversation
CREATE INDEX idx_messages_conversation ON messages(conversation_id, created_at DESC);
CREATE INDEX idx_messages_sender ON messages(sender_id);

-- Read status tracking
CREATE TABLE conversation_read_status (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_read_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (conversation_id, user_id)
);
