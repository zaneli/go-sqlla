DROP TABLE IF EXISTS "user";

CREATE TABLE "user" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name" TEXT NOT NULL,
    "age" INTEGER NULL,
    "rate" REAL NOT NULL DEFAULT 0,
    "created_at" DATETIME NOT NULL,
    "updated_at" DATETIME NULL
);
DROP TABLE IF EXISTS "user_item";

CREATE TABLE "user_item" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "user_id" INTEGER NOT NULL,
    "item_id" TEXT NOT NULL,
    "is_used" INTEGER NOT NULL,
    "has_extension" INTEGER NULL
);
