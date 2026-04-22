# 🤖 Go Discord Bot — LeetCode Daily Tracker

Discord bot viết bằng Go, tự động thông báo khi các thành viên trong server giải xong daily question LeetCode.

## ✨ Tính năng

| Command | Mô tả |
|---------|--------|
| `/register <username>` | Đăng ký LeetCode username để bot theo dõi |
| `/unregister` | Hủy đăng ký |
| `/stats [username]` | Xem thống kê LeetCode (solved, ranking, v.v.) |
| `/daily` | Xem daily challenge hôm nay |
| `/setchannel` | Set channel nhận thông báo (Admin only) |

Bot sẽ **tự động poll mỗi 2 phút** và gửi notification đẹp khi phát hiện user đã solve daily question.

## 🚀 Setup

### 1. Tạo Discord Bot

1. Vào [Discord Developer Portal](https://discord.com/developers/applications)
2. Tạo Application mới → lấy **Application ID**
3. Vào tab Bot → lấy **Bot Token**
4. Vào tab OAuth2 → URL Generator:
   - Scopes: `bot`, `applications.commands`
   - Permissions: `Send Messages`, `Embed Links`
5. Dùng URL đó để invite bot vào server

### 2. Cấu hình

```bash
cp .env.example .env
```

Sửa file `.env`:
```env
DISCORD_BOT_TOKEN=your_bot_token_here
DISCORD_APP_ID=your_app_id_here
DATABASE_PATH=./data/bot.db
POLL_INTERVAL=2m
LOG_LEVEL=info
```

### 3. Chạy

```bash
# Development
make dev

# Build & Run
make build
make run
```

## 🏗️ Kiến trúc

Project sử dụng **Hexagonal Architecture**:

```
cmd/bot/          → Entry point, dependency wiring
internal/
  domain/         → Entities & Port interfaces
  application/    → Use cases (business logic)
  adapter/
    inbound/      → Discord bot (slash commands)
    outbound/     → LeetCode API, SQLite storage
pkg/tracker/      → Background daily tracking worker
```

## 🛠️ Tech Stack

- **Go** + `discordgo` — Discord API
- **SQLite** (`modernc.org/sqlite`) — Storage (pure Go, no CGO)
- **LeetCode GraphQL API** — Data source (unofficial)
