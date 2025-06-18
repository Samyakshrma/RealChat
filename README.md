
---


# 🟢 RealChat – Golang WebSocket Chat Backend

RealChat is a lightweight real-time chat backend built in **Go** using:
- 🟣 Gin for HTTP routing
- 🧠 PostgreSQL (via pgx) for persistence
- 🔴 Redis Pub/Sub for real-time message delivery
- 🔐 JWT-based authentication

---

## 📦 Features

- ✅ User Registration & Login (hashed password + JWT)
- 💬 Real-time one-to-one chat with WebSocket
- 📨 Messages stored in PostgreSQL
- 🔐 Secure routes with JWT middleware

---

## 🚀 Tech Stack

| Component        | Description                      |
|------------------|----------------------------------|
| **Gin**          | Web framework (REST APIs)        |
| **pgx**          | PostgreSQL driver & connection pool |
| **Redis**        | Pub/Sub for WebSocket messaging  |
| **JWT (golang-jwt)** | Auth token handling            |
| **Gorilla WebSocket** | Bi-directional communication |

---

## 🛠️ Setup Instructions

### 1. Clone the repo
```bash
git clone https://github.com/Samyakshrma/RealChat.git
cd RealChat


### 2. Configure Environment

Create a `.env` file or export this before running:

```bash
export DATABASE_URL="postgres://<user>:<pass>@<host>:<port>/<dbname>?sslmode=require"
```

Make sure Redis is running locally on the default port (6379).

---

### 3. Database Schema

Run the following SQL to create the required tables:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    sender_id INT REFERENCES users(id),
    receiver_id INT REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 🧪 API Endpoints

### 🔐 Register

```
POST /register
```

**Body:**

```json
{
  "username": "alice",
  "password": "password123"
}
```

---

### 🔐 Login

```
POST /login
```

**Body:**

```json
{
  "username": "alice",
  "password": "password123"
}
```

**Response:**

```json
{
  "token": "<JWT_TOKEN>"
}
```

---

### 💬 WebSocket Chat

```
GET /chat?to=<target_user_id>
```

**Headers:**

```http
Authorization: Bearer <JWT_TOKEN>
```

Connect using WebSocket with messages like:

```json
{
  "to": "2",
  "content": "Hey there!"
}
```

---

## ▶️ Run the App

```bash
go run main.go
```

The server will run at `http://localhost:8080`

---

## 📁 Project Structure

```
.
├── main.go
├── config/         # DB & Redis setup
├── handlers/       # Register, Login, Chat endpoints
├── middleware/     # JWT auth middleware
├── models/         # User & Message models
├── utils/          # Context, Redis instance
```

---

## 📌 Notes

* Messages are saved in DB and also published via Redis Pub/Sub to the connected user's channel.
* JWT claims include `user_id` and `exp`, and are required for accessing `/chat`.

---

## 📬 Contact

Created by [@Samyakshrma](https://github.com/Samyakshrma) · MIT License

```

---

Let me know if you want:
- Dockerfile + docker-compose setup
- Swagger/OpenAPI spec
- Seed script for dev data
- VS Code devcontainer

I can also auto-generate `.env.example` and `Makefile` for local dev.
```
