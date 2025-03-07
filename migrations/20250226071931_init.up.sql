CREATE TYPE priority AS ENUM ('low', 'medium', 'high', 'urgent');
CREATE TYPE status AS ENUM ('planned', 'in_progress', 'completed', 'canceled');

CREATE TABLE IF NOT EXISTS todos (
    id SERIAL PRIMARY KEY,
    title text NOT NULL,
    description text,
    due_date DATE,
    tags text[],
    priority priority,
    status status,
    overdue bool
);