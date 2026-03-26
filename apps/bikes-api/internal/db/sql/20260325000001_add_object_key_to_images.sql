-- +goose Up
ALTER TABLE bike_images RENAME COLUMN url TO object_key;

-- +goose Down
ALTER TABLE bike_images RENAME COLUMN object_key TO url;
