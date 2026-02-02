-- 公告通知功能
-- 用于管理员发布公告，用户登录后可以弹窗查看

-- 公告表
CREATE TABLE IF NOT EXISTS announcements (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'info',      -- info/warning/important
    priority INT NOT NULL DEFAULT 0,               -- 优先级，越大越优先显示
    status VARCHAR(20) NOT NULL DEFAULT 'draft',   -- draft/published/archived
    publish_at TIMESTAMPTZ,                        -- 定时发布时间，NULL 表示立即发布
    expires_at TIMESTAMPTZ,                        -- 过期时间，NULL 表示永不过期
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ                         -- 软删除
);

-- 用户已读记录表
CREATE TABLE IF NOT EXISTS announcement_reads (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    announcement_id BIGINT NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, announcement_id)
);

-- 公告表索引
CREATE INDEX IF NOT EXISTS idx_announcements_status ON announcements(status);
CREATE INDEX IF NOT EXISTS idx_announcements_publish_at ON announcements(publish_at);
CREATE INDEX IF NOT EXISTS idx_announcements_expires_at ON announcements(expires_at);
CREATE INDEX IF NOT EXISTS idx_announcements_deleted_at ON announcements(deleted_at);
CREATE INDEX IF NOT EXISTS idx_announcements_priority ON announcements(priority DESC);
CREATE INDEX IF NOT EXISTS idx_announcements_created_at ON announcements(created_at DESC);

-- 已读记录表索引
CREATE INDEX IF NOT EXISTS idx_announcement_reads_user_id ON announcement_reads(user_id);
CREATE INDEX IF NOT EXISTS idx_announcement_reads_announcement_id ON announcement_reads(announcement_id);

-- 表注释
COMMENT ON TABLE announcements IS '公告表';
COMMENT ON COLUMN announcements.title IS '公告标题';
COMMENT ON COLUMN announcements.content IS '公告内容（支持 Markdown）';
COMMENT ON COLUMN announcements.type IS '公告类型: info-普通, warning-警告, important-重要';
COMMENT ON COLUMN announcements.priority IS '优先级，数值越大越优先显示';
COMMENT ON COLUMN announcements.status IS '状态: draft-草稿, published-已发布, archived-已归档';
COMMENT ON COLUMN announcements.publish_at IS '定时发布时间';
COMMENT ON COLUMN announcements.expires_at IS '过期时间';
COMMENT ON COLUMN announcements.created_by IS '创建者用户ID';

COMMENT ON TABLE announcement_reads IS '公告已读记录表';
COMMENT ON COLUMN announcement_reads.user_id IS '用户ID';
COMMENT ON COLUMN announcement_reads.announcement_id IS '公告ID';
COMMENT ON COLUMN announcement_reads.read_at IS '阅读时间';
