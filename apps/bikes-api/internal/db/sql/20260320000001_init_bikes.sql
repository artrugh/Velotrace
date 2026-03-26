-- +goose Up
CREATE TABLE IF NOT EXISTS bikes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    make_model TEXT NOT NULL,
    year INTEGER,
    price DECIMAL(10,2),
    location_city TEXT,
    current_owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    serial_number TEXT UNIQUE NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'registered',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS bike_images (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bike_id UUID NOT NULL REFERENCES bikes(id) ON DELETE CASCADE,
    object_key TEXT NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ownership_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bike_id UUID NOT NULL REFERENCES bikes(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_active BOOLEAN DEFAULT TRUE,
    acquired_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    sold_at TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE IF EXISTS ownership_records;
DROP TABLE IF EXISTS bike_images;
DROP TABLE IF EXISTS bikes;
