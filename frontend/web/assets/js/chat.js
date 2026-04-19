const params = new URLSearchParams(window.location.search);
const room = params.get("room");
const username = params.get("username");
const useAI = params.get("useAI") === "true";

if (!room || !username) {
  alert("Room or username missing. Redirecting to homepage...");
  window.location.href = "/";
}

const socket = new WebSocket(
  `ws://${location.host}/room?room=${encodeURIComponent(room)}&name=${encodeURIComponent(username)}&useAI=${useAI}`
);

// Store reference the active streaming bubbles
const streamingBubbles = {};

socket.onmessage = (event) => {
  try {
    const data = JSON.parse(event.data);
    const messagesDiv = document.getElementById("messages");

    if (data.streamId && data.streaming !== undefined && data.streaming === true) {

      if (streamingBubbles[data.streamId]) {
        // Add token
        streamingBubbles[data.streamId].textContent += data.message;

      } else {
        const msgContainer = document.createElement("div");
        msgContainer.classList.add("message-container");

        const usernameDiv = document.createElement("div");
        usernameDiv.classList.add("username");
        usernameDiv.textContent = data.name;

        const messageDiv = document.createElement("div");
        messageDiv.classList.add("message", "streaming");
        messageDiv.textContent = data.message;

        msgContainer.appendChild(usernameDiv);
        msgContainer.appendChild(messageDiv);
        messagesDiv.appendChild(msgContainer);

        // Store reference
        streamingBubbles[data.streamId] = messageDiv;
      }

    }else if (data.streamId && data.streaming !== undefined && data.streaming === false) {
      if (streamingBubbles[data.streamId]) {
        streamingBubbles[data.streamId].textContent = data.message;
        streamingBubbles[data.streamId].classList.remove("streaming");
        delete streamingBubbles[data.streamId];
      }

    } else {
      const msgContainer = document.createElement("div");
      msgContainer.classList.add("message-container");

      const usernameDiv = document.createElement("div");
      usernameDiv.classList.add("username");
      usernameDiv.textContent = data.name;

      const messageDiv = document.createElement("div");
      messageDiv.classList.add("message");
      messageDiv.textContent = data.message;

      msgContainer.appendChild(usernameDiv);
      msgContainer.appendChild(messageDiv);
      messagesDiv.appendChild(msgContainer);
    }

    // Auto-scroll
    messagesDiv.scrollTop = messagesDiv.scrollHeight;

  } catch (err) {
    console.error("Invalid JSON received:", event.data);
  }
};

function sendMessage() {
  const input = document.getElementById("msg");
  if (input.value.trim() !== "") {
    socket.send(input.value);
    input.value = "";
  }
}

document.getElementById("sendBtn").addEventListener("click", sendMessage);

document.getElementById("msg").addEventListener("keyup", function (event) {
  if (event.key === "Enter") {
    sendMessage();
  }
});