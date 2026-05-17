async function login() {
  const email = document.getElementById("email").value;
  const password = document.getElementById("password").value;
  const error = document.getElementById("error");

  error.textContent = "";

  try {
    const response = await fetch("/api/auth/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      credentials: "include",
      body: JSON.stringify({
        email,
        password
      })
    });

    const data = await response.json();

    if (!response.ok) {
      error.textContent = data.message || "Invalid credentials";
      return;
    }

    window.location.href = "/";

  } catch (err) {
    error.textContent = "Network error";
  }
}

async function logout() {
  await fetch("/api/auth/logout", {
    method: "GET",
    credentials: "include"
  });

  window.location.href = "/login";
}

async function alreadyLoggedIn() {
try {
  const response = await fetch("/api/auth/me", {
    credentials: "include"
  });

  if (response.ok) {
    window.location.href = "/";
  }
} catch {}
}

async function signup() {
  const name = document.getElementById("name").value;
  const email = document.getElementById("email").value;
  const password = document.getElementById("password").value;
  const error = document.getElementById("error");

  error.textContent = "";

  try {
    const response = await fetch("/api/auth/signup", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      credentials: "include",
      body: JSON.stringify({
        name,
        email,
        password
      })
    });

    const data = await response.json();

    if (!response.ok) {
      error.textContent = data.message || "Signup failed";
      return;
    }

    // auto login after signup
    const loginResponse = await fetch("/api/auth/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      credentials: "include",
      body: JSON.stringify({
        email,
        password
      })
    });

    if (!loginResponse.ok) {
      error.textContent = "User created, but login failed";
      return;
    }

    window.location.href = "/";

  } catch (err) {
    error.textContent = "Network error";
  }
}

 
