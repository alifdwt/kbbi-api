-- Create dictionary table
CREATE TABLE IF NOT EXISTS public."dictionary" (
    id SERIAL PRIMARY KEY,
    word VARCHAR(255) NOT NULL,
    arti TEXT NOT NULL,
    "type" INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Enable pg_trgm extension for better text search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_dictionary_word ON public."dictionary"(word);
CREATE INDEX IF NOT EXISTS idx_dictionary_word_gin ON public."dictionary" USING gin(word gin_trgm_ops);

-- Update updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_dictionary_updated_at
    BEFORE UPDATE ON public."dictionary"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();