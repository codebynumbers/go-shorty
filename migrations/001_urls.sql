CREATE TABLE `urls` (
    `hash` VARCHAR (10) PRIMARY KEY,
    `url` VARCHAR(2048) NOT NULL
);
CREATE UNIQUE INDEX idx_url ON urls (url);
