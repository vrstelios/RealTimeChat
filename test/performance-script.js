/**
 * RealTimeChat — Senior Load Test
 * ================================
 * Tests all system capabilities:
 *  1. Auth (signup + login)
 *  2. WebSocket chat (multiple concurrent users)
 *  3. AI responses (Gemini streaming)
 *  4. PDF upload + RAG query
 *  5. Multi-room isolation
 *
 * Run: node load-test.js
 * Requirements: npm install ws node-fetch form-data
 */

const WebSocket = require("ws");
const FormData = require("form-data");
const fs = require("fs");
const path = require("path");

const fetch = globalThis.fetch || require("node-fetch");

// ── CONFIG ──────────────────────────────────────────────
const BASE_URL = "http://localhost:8080";
const WS_URL   = "ws://localhost:8080";

const CONFIG = {
  CONCURRENT_USERS: 10,       // users per room
  ROOMS: ["room-load-1", "room-load-2", "room-load-3", "room-load-4", "room-load-5"],
  MESSAGES_PER_USER: 10,      // messages each user sends
  MESSAGE_DELAY_MS: 500,     // delay between messages
  AI_ROOM: "room-ai-load",   // room for AI testing
  AI_QUESTIONS: [
    "What is a WebSocket?",
    "How does Redis Pub/Sub work?",
    "Explain horizontal scaling",
  ],
};

// ── STATS ────────────────────────────────────────────────
const stats = {
  signups: { success: 0, failed: 0 },
  logins: { success: 0, failed: 0 },
  messages: { sent: 0, received: 0 },
  websockets: { connected: 0, errors: 0 },
  ai: { requests: 0, success: 0, failed: 0 },
  uploads: { success: 0, failed: 0 },
  startTime: Date.now(),
};

// ── HELPERS ──────────────────────────────────────────────
function log(emoji, label, msg) {
  const elapsed = ((Date.now() - stats.startTime) / 1000).toFixed(1);
  console.log(`[${elapsed}s] ${emoji}  ${label.padEnd(20)} ${msg}`);
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function randomString(len = 8) {
  return Math.random().toString(36).substring(2, 2 + len);
}

// ── AUTH ─────────────────────────────────────────────────
async function signup(name, email, password) {
  try {
    const res = await fetch(`${BASE_URL}/api/auth/signup`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, email, password }),
    });

    const body = await res.text();
    log(" ", "Signup", `${email} → ${res.status}: ${body.substring(0, 100)}`);

    if (res.ok) {
      stats.signups.success++;
      return true;
    }
    stats.signups.failed++;
    return false;
  } catch (err) {
    log(" ", "Signup Error", err.message);
    stats.signups.failed++;
    return false;
  }
}

async function login(email, password) {
  try {
    const res = await fetch(`${BASE_URL}/api/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });

    if (res.ok) {
      stats.logins.success++;
      // Extract cookie
      const setCookie = res.headers.get("set-cookie") || "";
      const match = setCookie.match(/Authorization=([^;]+)/);
      return match ? match[1] : null;
    }
    stats.logins.failed++;
    return null;
  } catch {
    stats.logins.failed++;
    return null;
  }
}

// ── WEBSOCKET USER ───────────────────────────────────────
function createUser(username, room, token, useAI = false) {
  return new Promise((resolve) => {
    const url = `${WS_URL}/api/room?room=${encodeURIComponent(room)}&name=${encodeURIComponent(username)}&useAI=${useAI}&token=${token}`;
    const ws = new WebSocket(url);

    let connected = false;
    let messagesReceived = 0;

    ws.on("open", () => {
      connected = true;
      stats.websockets.connected++;
      log(" ", "WS Connected", `${username} → ${room}`);
      resolve({ ws, username, room, getReceived: () => messagesReceived });
    });

    ws.on("message", (data) => {
      messagesReceived++;
      stats.messages.received++;

      try {
        const msg = JSON.parse(data.toString());
        // Log AI streaming completion
        if (msg.name === "Gemini" && msg.streaming === false) {
          stats.ai.success++;
          log(" ", "AI Response", `[${room}] ${msg.message.substring(0, 60)}...`);
        }
      } catch {}
    });

    ws.on("error", (err) => {
      stats.websockets.errors++;
      if (!connected) resolve(null);
    });

    ws.on("close", () => {
      if (!connected) resolve(null);
    });

    // Timeout
    setTimeout(() => {
      if (!connected) {
        stats.websockets.errors++;
        resolve(null);
      }
    }, 5000);
  });
}

// ── SEND MESSAGES ────────────────────────────────────────
async function sendMessages(user, messages) {
  if (!user) return;
  const { ws, username, room } = user;

  for (const msg of messages) {
    if (ws.readyState !== WebSocket.OPEN) break;
    ws.send(msg);
    stats.messages.sent++;
    log(" ", "Message Sent", `${username} → [${room}]: "${msg}"`);
    await sleep(CONFIG.MESSAGE_DELAY_MS);
  }
}

// ── PDF UPLOAD ───────────────────────────────────────────
async function uploadPDF(room, token) {
  try {
    const pdfPath = path.join(__dirname, "TestRealTimeChat.pdf");

    log(" ", "PDF Path", pdfPath);
    log(" ", "PDF Exists", fs.existsSync(pdfPath).toString());

    if (!fs.existsSync(pdfPath)) {
      log(" ", "PDF Upload", "File not found, skipping");
      stats.uploads.failed++;
      return false;
    }

    const form = new globalThis.FormData();
    const fileBuffer = fs.readFileSync(pdfPath);
    const fileBlob = new Blob([fileBuffer], { type: "application/pdf" });
    form.append("file", fileBlob, "TestRealTimeChat.pdf");

    const res = await fetch(
      `${BASE_URL}/api/documents/upload?room=${encodeURIComponent(room)}`,
      {
        method: "POST",
        headers: {
          Cookie: `Authorization=${token}`,
        },
        body: form,
      }
    );

    const body = await res.text();
    log(" ", "PDF Upload", `${res.status}: ${body.substring(0, 150)}`);

    if (res.ok) {
      stats.uploads.success++;
      return true;
    }
    stats.uploads.failed++;
    return false;
  } catch (err) {
    stats.uploads.failed++;
    log(" ", "PDF Upload", `Error: ${err.message}`);
    return false;
  }
}

// ── TEST SCENARIOS ───────────────────────────────────────

// Scenario 1: Multi-user chat in multiple rooms
async function scenarioChatLoad() {
  log(" ", "SCENARIO 1", "Multi-user chat load test");

  // Create test users
  const users = [];
  for (let i = 0; i < CONFIG.CONCURRENT_USERS; i++) {
    const id = randomString(6);
    const email = `loadtest-${id}@test.com`;
    const password = "test1234";
    const name = `user-${id}`;

    await signup(name, email, password);
    const token = await login(email, password);
    if (token) users.push({ name, email, password, token });
  }

  log(" ", "Users Created", `${users.length}/${CONFIG.CONCURRENT_USERS} ready`);

  // Connect users to rooms
  const connections = await Promise.all(
    users.map((u, i) => {
      const room = CONFIG.ROOMS[i % CONFIG.ROOMS.length];
      return createUser(u.name, room, u.token);
    })
  );

  await sleep(1000);

  // Send messages concurrently
  const messages = Array.from(
    { length: CONFIG.MESSAGES_PER_USER },
    (_, i) => `Load test message ${i + 1} — ${randomString(10)}`
  );

  await Promise.all(
    connections
      .filter(Boolean)
      .map((conn) => sendMessages(conn, messages))
  );

  await sleep(2000);

  // Close connections
  connections.filter(Boolean).forEach(({ ws }) => ws.close());

  log(" ", "SCENARIO 1", "Complete");
  return users;
}

// Scenario 2: AI interaction test
async function scenarioAILoad(users) {
  log(" ", "SCENARIO 2", "AI interaction test");

  if (!users || users.length === 0) {
    log(" ", "SCENARIO 2", "No users available, skipping");
    return;
  }

  const user = users[0];
  const conn = await createUser(user.name, CONFIG.AI_ROOM, user.token, true);

  if (!conn) {
    log(" ", "SCENARIO 2", "Failed to connect");
    return;
  }

  await sleep(500);

  for (const question of CONFIG.AI_QUESTIONS) {
    stats.ai.requests++;
    log(" ", "AI Question", question);
    conn.ws.send(question);
    await sleep(8000); // Αναμονή για AI response
  }

  conn.ws.close();
  log(" ", "SCENARIO 2", "Complete");
}

// Scenario 3: PDF upload + RAG query
async function scenarioRAGLoad(users) {
  log(" ", "SCENARIO 3", "PDF Upload + RAG query test");

  if (!users || users.length === 0) {
    log(" ", "SCENARIO 3", "No users available, skipping");
    return;
  }

  const user = users[0];
  const ragRoom = "room-rag-load";

  // Upload PDF
  const uploaded = await uploadPDF(ragRoom, user.token);
  if (!uploaded) {
    log(" ", "SCENARIO 3", "Upload failed, skipping RAG query");
    return;
  }

  await sleep(2000); // Wait for embedding

  // Connect and query
  const conn = await createUser(user.name, ragRoom, user.token, true);
  if (!conn) return;

  await sleep(500);

  conn.ws.send("Who is Stylianos Verros according to the document?");
  stats.ai.requests++;
  log(" ", "RAG Query", "Asking about document content...");

  await sleep(10000); // Wait for RAG + AI response

  conn.ws.close();
  log(" ", "SCENARIO 3", "Complete");
}

// Scenario 4: Concurrent room isolation test
async function scenarioRoomIsolation(users) {
  log(" ", "SCENARIO 4", "Room isolation test");

  if (!users || users.length < 2) {
    log(" ", "SCENARIO 4", "Not enough users, skipping");
    return;
  }

  // Two users in different rooms — messages should NOT cross
  const [u1, u2] = users;

  const [conn1, conn2] = await Promise.all([
    createUser(u1.name, "isolated-room-A", u1.token),
    createUser(u2.name, "isolated-room-B", u2.token),
  ]);

  await sleep(500);

  let crossRoomMessage = false;

  if (conn1) {
    conn1.ws.on("message", (data) => {
      const msg = JSON.parse(data.toString());
      if (msg.name === u2.name) {
        crossRoomMessage = true;
        log(" ", "Room Isolation", "FAIL — Message crossed room boundary!");
      }
    });
  }

  if (conn1) conn1.ws.send("Message in Room A — should NOT appear in Room B");
  if (conn2) conn2.ws.send("Message in Room B — should NOT appear in Room A");

  await sleep(2000);

  if (!crossRoomMessage) {
    log(" ", "Room Isolation", "PASS — Rooms are properly isolated");
  }

  if (conn1) conn1.ws.close();
  if (conn2) conn2.ws.close();

  log(" ", "SCENARIO 4", "Complete");
}

// ── PRINT STATS ──────────────────────────────────────────
function printStats() {
  const elapsed = ((Date.now() - stats.startTime) / 1000).toFixed(1);

  console.log("\n" + "═".repeat(55));
  console.log("LOAD TEST RESULTS");
  console.log("═".repeat(55));
  console.log(`Duration: ${elapsed}s`);
  console.log("─".repeat(55));
  console.log("AUTH");
  console.log(`Signups: ${stats.signups.success} ${stats.signups.failed}`);
  console.log(`Logins: ${stats.logins.success} ${stats.logins.failed}`);
  console.log("─".repeat(55));
  console.log("WEBSOCKET");
  console.log(`Connected: ${stats.websockets.connected} ${stats.websockets.errors}`);
  console.log(`Messages Sent:${stats.messages.sent}`);
  console.log(`Messages Recv:${stats.messages.received}`);
  console.log("─".repeat(55));
  console.log("AI (GEMINI)");
  console.log(`Requests: ${stats.ai.requests}`);
  console.log(`Success: ${stats.ai.success} ${stats.ai.failed}`);
  console.log("─".repeat(55));
  console.log("DOCUMENTS (RAG)");
  console.log(`Uploads: ${stats.uploads.success} ${stats.uploads.failed}`);
  console.log("─".repeat(55));
  console.log(`Throughput:${(stats.messages.sent / elapsed).toFixed(1)} msg/s`);
  console.log("═".repeat(55));
  console.log("\n Check Grafana: http://localhost:3000");
  console.log(" Check Jaeger:  http://localhost:16686\n");
}

// ── MAIN ─────────────────────────────────────────────────
async function main() {
  console.log("\n" + "═".repeat(55));
  console.log("  ⬡  RealTimeChat Load Test");
  console.log("  Senior-level system demonstration");
  console.log("═".repeat(55) + "\n");

  try {
    // Run scenarios sequentially
    const users = await scenarioChatLoad();
    await sleep(2000);

    await scenarioAILoad(users);
    await sleep(2000);

    await scenarioRAGLoad(users);
    await sleep(2000);

    await scenarioRoomIsolation(users);
    await sleep(2000);

  } catch (err) {
    console.error("Test error:", err);
  } finally {
    printStats();
    process.exit(0);
  }
}

main();