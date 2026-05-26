const params = new URLSearchParams(window.location.search);
const room = params.get("room");
const username = params.get("username");
const useAI = params.get("useAI") === "true";
const protocol = location.protocol === "https:" ? "wss" : "ws";

console.log("[Chat Initialized]", { room, username, useAI, protocol });

if (!room || !username) {
  alert("Room or username missing. Redirecting to homepage...");
  window.location.href = "/";
}

// Update UI Text elements
document.getElementById("roomName").textContent = `# ${room}`;
document.getElementById("headerRoomTitle").textContent = room;
document.getElementById("currentUsername").textContent = username;
if(!useAI) {
  console.log("[AI Status] Disabled via URL parameters.");
  document.getElementById("aiStatus").style.display = 'none';
}

function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return '';
}

const token = getCookie('Authorization');

const wsUrl = `${protocol}://${location.host}/api/room?room=${encodeURIComponent(room)}&name=${encodeURIComponent(username)}&useAI=${useAI}&token=${token}`;
console.log("[WebSocket Connecting to]:", wsUrl);

const socket = new WebSocket(wsUrl);

const streamingBubbles = {};
const typingIndicator = document.getElementById("typingIndicator");

socket.onopen = () => {
  console.log("[WebSocket] Connection established successfully.");
};

socket.onerror = (error) => {
  console.error("[WebSocket Error]:", error);
};

socket.onclose = (event) => {
  console.warn("[WebSocket Closed]:", event);
};

// Handle incoming messages
socket.onmessage = (event) => {
  try {
    const data = JSON.parse(event.data);
    const messagesDiv = document.getElementById("messages");

    console.log("[WebSocket Received Data]:", data);

    // Live typing status notification
    if (data.type === "typing") {
       showTypingIndicator(`${data.name} is typing...`);
       return;
    }

    if (data.streamId && data.streaming !== undefined && data.streaming === true) {
      // AI Typing / Streaming Flow
      console.log(`[AI Streaming Chunk] StreamID: ${data.streamId}, Message snippet: ${data.message}`);
      showTypingIndicator("Gemini AI is thinking...");

      if (streamingBubbles[data.streamId]) {
        streamingBubbles[data.streamId].textContent += data.message;
      } else {
        console.log(`[AI Stream Start] Creating bubble for StreamID: ${data.streamId}`);
        createMessageBubble(data.name, data.message, data.streamId, true);
      }
    } else if (data.streamId && data.streaming !== undefined && data.streaming === false) {
      // End of AI Stream
      console.log(`[AI Stream End] Finished StreamID: ${data.streamId}`);
      hideTypingIndicator();
      if (streamingBubbles[data.streamId]) {
        streamingBubbles[data.streamId].textContent = data.message;
        streamingBubbles[data.streamId].classList.remove("streaming");
        delete streamingBubbles[data.streamId];
      }
    } else {
      // Standard Chat Message
      console.log(`[Standard Message] From: ${data.name}`);
      createMessageBubble(data.name, data.message, null, false);
    }

    // Auto-scroll
    messagesDiv.scrollTop = messagesDiv.scrollHeight;
  } catch (err) {
    console.error("[JSON Parse Error] Raw data received:", event.data, err);
  }
};

function createMessageBubble(sender, text, streamId, isStreaming) {
  const messagesDiv = document.getElementById("messages");
  const msgContainer = document.createElement("div");
  msgContainer.classList.add("message-container");

  // Style layout according to sender
  if (sender === username) {
    msgContainer.classList.add("my-message");
  } else if (sender.toLowerCase().includes("gemini") || sender.toLowerCase().includes("ai")) {
    msgContainer.classList.add("ai-message");
  } else {
    msgContainer.classList.add("other-message");
  }

  const usernameDiv = document.createElement("div");
  usernameDiv.classList.add("username");
  usernameDiv.textContent = sender;

  const messageDiv = document.createElement("div");
  messageDiv.classList.add("message");
  if (isStreaming) {
    messageDiv.classList.add("streaming");
  }
  messageDiv.textContent = text;

  msgContainer.appendChild(usernameDiv);
  msgContainer.appendChild(messageDiv);
  messagesDiv.appendChild(msgContainer);

  if (isStreaming && streamId) {
    streamingBubbles[streamId] = messageDiv;
    console.log(`[DOM Referenced] Registered stream bubble in tracking array for ID: ${streamId}`);
  }
}

function showTypingIndicator(text) {
  typingIndicator.textContent = text;
  typingIndicator.classList.add("active");
}

function hideTypingIndicator() {
  typingIndicator.classList.remove("active");
}

// Send Message Logic
function sendMessage() {
  const input = document.getElementById("msg");
  const messageText = input.value.trim();

  if (messageText !== "") {
    console.log("[Sending Message payload via WS]:", messageText);
    socket.send(messageText);
    input.value = "";
    hideTypingIndicator();
  }
}

document.getElementById("sendBtn").addEventListener("click", (e) => {
  e.preventDefault();
  sendMessage();
});

document.getElementById("msg").addEventListener("keyup", (e) => {
  if (e.key === "Enter") {
    sendMessage();
  }
});

// Professional File Upload System (RAG Engine Connection)
const fileInput = document.getElementById('fileInput');
const uploadBtn = document.getElementById('uploadBtn');
const uploadPreview = document.getElementById('uploadPreview');
const fileNameDisplay = document.getElementById('fileNameDisplay');
const cancelUpload = document.getElementById('cancelUpload');

uploadBtn.addEventListener('click', (e) => {
    e.preventDefault();
    fileInput.click();
});

fileInput.addEventListener('change', function() {
    if (this.files.length === 0) return;
    const file = this.files[0];

    console.log("[File Selected for Upload]:", file.name, file.type, file.size);

    if (file.type !== "application/pdf") {
        alert("Only high-grade PDF documents are accepted for RAG injection.");
        this.value = '';
        return;
    }

    // Display professional preview bar
    fileNameDisplay.textContent = `${file.name} (${(file.size/1024/1024).toFixed(2)} MB)`;
    uploadPreview.style.display = 'flex';

    // Auto-execute RAG pipeline
    executeUpload(file);
});

cancelUpload.addEventListener('click', () => {
    console.log("[Upload Cancelled by User]");
    uploadPreview.style.display = 'none';
    fileInput.value = '';
});

async function executeUpload(file) {
    const formData = new FormData();
    formData.append('file', file);

    uploadBtn.classList.add("loading");
    fileNameDisplay.textContent = `Injecting into Vector DB... ⏳`;

    const uploadUrl = `/api/documents/upload?room=${encodeURIComponent(room)}`;
    console.log("[HTTP POST Uploading to]:", uploadUrl);

    try {
        const response = await fetch(uploadUrl, {
            method: 'POST',
            body: formData,
            credentials: 'include'
        });
        const result = await response.json();

        console.log("[HTTP Upload Server Response]:", response.status, result);

        if (response.ok) {
            fileNameDisplay.textContent = `🚀 Success! Document indexed successfully.`;
            setTimeout(() => { uploadPreview.style.display = 'none'; }, 3000);
        } else {
            alert("RAG Engine Refused Document: " + (result.message || "Unknown Error"));
            uploadPreview.style.display = 'none';
        }
    } catch (error) {
        console.error("[Network Error during PDF upload]:", error);
        alert("Network Error while injection data into Qdrant.");
        uploadPreview.style.display = 'none';
    } finally {
        uploadBtn.classList.remove("loading");
        fileInput.value = ''; // Clean input placeholder correctly
    }
}