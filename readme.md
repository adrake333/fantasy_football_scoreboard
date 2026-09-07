# 🏈 Fantasy Football Live Scoreboard

A real-time, multi-user fantasy football dashboard built in Go. It aggregates leagues and live matchups across both **Sleeper** and **ESPN Fantasy Football**, streaming live score updates directly to the browser via Server-Sent Events (SSE).

---

## ✨ Features

- **Multi-Platform Support**: Aggregates leagues from both Sleeper and ESPN into a single, unified scoreboard.
- **Real-Time Live Streaming**: Uses Server-Sent Events (SSE) with ticker intervals to stream live score and player stat updates without refreshing the page.
- **Multi-User Authentication**: Session-based auth with bcrypt password hashing and SQLite persistence.
- **Isolated User Data**: Each user registers, connects their respective Sleeper / ESPN credentials, and manages their own active leagues.
- **Lightweight & Fast**: Built with pure Go, standard library HTTP routing, and SQLite powered by `sqlc`.

---

## 🛠️ Tech Stack

- **Backend**: Go (Golang)
- **Database**: SQLite with [`sqlc`](https://sqlc.dev/) for type-safe query generation
- **Auth**: Bcrypt password hashing + secure HTTP-only session cookies
- **Frontend**: Server-rendered HTML templates + real-time JavaScript `EventSource` (SSE)
- **External APIs**: Sleeper REST API & ESPN Fantasy v3 Private API

---

## 🚀 Getting Started

### Prerequisites

- [Go 1.22+](https://golang.org/dl/)
- SQLite3 (optional, for inspecting the DB directly)
- `sqlc` (optional, for regenerating queries)

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/adrake333/fantasy_football_scoreboard.git
   cd fantasy_football_scoreboard

2. **Download dependencies:**
    go mod download

3. **Run the application:**
    go run .
    
4. **Open your browser:**
    Navigate to http://localhost:8080

### Usage Guide

1. **Register an account:**
    Go to /register and create a username and password.
    Enter your Sleeper User ID and/or you ESPN SWID and espn_s2 cookies.

2. **Add leagues:**
    Add your active Sleeper or ESPN league IDs

3. **Watch the scores roll in:**
    The Dashboard will automatically filter to your active team's matchups and push live updates every 15 seconds!

### License

    MIT License. Feel free to use and modify for your own fantasy leagues!