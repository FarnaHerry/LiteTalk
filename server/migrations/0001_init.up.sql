-- 0001_init.up.sql — LiteTalk 核心表（落地方案 §6.1）
CREATE TABLE users (
    id            BIGINT       NOT NULL AUTO_INCREMENT,
    email         VARCHAR(255) NOT NULL,
    nickname      VARCHAR(16)  NOT NULL,
    tag           CHAR(4)      NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email_verified TINYINT(1)  NOT NULL DEFAULT 0,
    status        TINYINT      NOT NULL DEFAULT 0,
    created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_users_email (email),
    UNIQUE KEY uk_users_nickname_tag (nickname, tag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE friendships (
    id         BIGINT   NOT NULL AUTO_INCREMENT,
    user_id    BIGINT   NOT NULL,
    friend_id  BIGINT   NOT NULL,
    status     ENUM('pending','accepted','blocked') NOT NULL DEFAULT 'pending',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_friendships_pair (user_id, friend_id),
    KEY idx_friendships_friend (friend_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE servers (
    id         BIGINT       NOT NULL AUTO_INCREMENT,
    name       VARCHAR(64)  NOT NULL,
    owner_id   BIGINT       NOT NULL,
    icon_name  VARCHAR(64)  NOT NULL DEFAULT 'messages-square',
    icon_color VARCHAR(16)  NOT NULL DEFAULT '#4FA4F3',
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_servers_owner (owner_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE server_members (
    server_id BIGINT      NOT NULL,
    user_id   BIGINT      NOT NULL,
    role      ENUM('owner','admin','member') NOT NULL DEFAULT 'member',
    joined_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (server_id, user_id),
    KEY idx_server_members_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE channels (
    id         BIGINT      NOT NULL AUTO_INCREMENT,
    server_id  BIGINT      NOT NULL,
    name       VARCHAR(64) NOT NULL,
    type       ENUM('text','voice') NOT NULL DEFAULT 'text',
    position   INT         NOT NULL DEFAULT 0,
    created_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_channels_server (server_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE messages (
    id         BIGINT   NOT NULL AUTO_INCREMENT,
    channel_id BIGINT   NOT NULL,
    user_id    BIGINT   NOT NULL,
    content    TEXT     NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_messages_channel_id (channel_id, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE invites (
    code       VARCHAR(32) NOT NULL,
    server_id  BIGINT      NOT NULL,
    creator_id BIGINT      NOT NULL,
    max_uses   INT         NOT NULL DEFAULT 0,
    uses       INT         NOT NULL DEFAULT 0,
    expires_at DATETIME    NULL,
    created_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (code),
    KEY idx_invites_server (server_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
