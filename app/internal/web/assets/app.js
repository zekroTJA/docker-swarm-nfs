"use strict";

const REFRESH_INTERVAL_MS = 3000;

const entriesList = document.getElementById("entries");
const form = document.getElementById("entry-form");
const nameInput = document.getElementById("name-input");
const messageInput = document.getElementById("message-input");
const submitButton = form.querySelector("button");
const formError = document.getElementById("form-error");

// ID of the newest entry currently rendered. Used as `last_id` so the server
// only returns entries we have not seen yet.
let lastId = "";

async function fetchNewEntries() {
  const query = lastId ? `?last_id=${encodeURIComponent(lastId)}` : "";
  const response = await fetch(`/api/entries${query}`);
  if (!response.ok) {
    throw new Error(`unexpected status ${response.status}`);
  }
  return response.json();
}

function formatTimestamp(value) {
  return new Date(value).toLocaleString();
}

function renderEntry(entry) {
  const item = document.createElement("li");
  item.className = "entry is-new";

  const header = document.createElement("div");
  header.className = "entry-header";

  const name = document.createElement("span");
  name.className = "entry-name";
  name.textContent = entry.name;

  const time = document.createElement("time");
  time.className = "entry-time";
  time.dateTime = entry.createdAt;
  time.textContent = formatTimestamp(entry.createdAt);

  header.append(name, time);

  const message = document.createElement("p");
  message.className = "entry-message";
  message.textContent = entry.message;

  item.append(header, message);
  return item;
}

function prependEntries(entries) {
  // The API returns entries oldest first. Prepending them in that order leaves
  // the list sorted newest first, with the latest entry at the very top.
  for (const entry of entries) {
    entriesList.prepend(renderEntry(entry));
    lastId = entry.id;
  }
}

async function refresh() {
  try {
    prependEntries(await fetchNewEntries());
  } catch (error) {
    console.error("failed refreshing entries", error);
  }
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  formError.textContent = "";
  submitButton.disabled = true;

  try {
    const response = await fetch("/api/entries", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: nameInput.value.trim(),
        message: messageInput.value.trim(),
      }),
    });

    if (!response.ok) {
      const body = await response.json().catch(() => ({}));
      throw new Error(body.error || `unexpected status ${response.status}`);
    }

    messageInput.value = "";
    await refresh();
  } catch (error) {
    formError.textContent = error.message;
  } finally {
    submitButton.disabled = false;
  }
});

refresh();
setInterval(refresh, REFRESH_INTERVAL_MS);
