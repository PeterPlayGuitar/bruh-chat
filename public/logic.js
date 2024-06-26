url = "http://rabit1998g.fvds.ru:65000";

document.addEventListener("DOMContentLoaded", () => {
  const textarea = document.getElementById("input");
  const senderInput = document.getElementById("senderInput");
  const messagesBox = document.getElementById("messages");

  everyPeriod(1000, () => {
    update(messagesBox);
  });

  textarea.addEventListener("input", function (e) {
    this.style.height = "auto";
    if (this.scrollHeight < 200) this.style.height = this.scrollHeight + "px";
    else this.style.height = "200px";
  });

  textarea.addEventListener("keydown", function (event) {
    if (event.key === "Enter" || event.keyCode === 13) {
      event.preventDefault();

      if (textarea.value === "") return;

      sendMessage(
        {
          value: textarea.value,
          sender: senderInput.value === "" ? "anonim" : senderInput.value,
        },
        () => {
          textarea.value = "";
          update(messagesBox);
        }
      );
    }
  });

  update(messagesBox);
});

function everyPeriod(time, func) {
  setTimeout(() => {
    func();
    everyPeriod(time, func);
  }, time);
}

function update(messagesBox) {
  getMessages((response) => {
    messagesBox.innerHTML = "";
    if (response === null) return;
    response.reverse().forEach((element) => {
      var msgDiv = document.createElement("div");
      var valueDiv = document.createElement("div");
      var senderDiv = document.createElement("div");

      // Add content and attributes to the div
      valueDiv.textContent = element.value;
      senderDiv.textContent = element.sender;
      msgDiv.classList.add("message");
      senderDiv.classList.add("sender");
      valueDiv.classList.add("value");
      msgDiv.appendChild(senderDiv);
      msgDiv.appendChild(valueDiv);

      // Append the new div to the body (or any other parent element)
      messagesBox.appendChild(msgDiv);
    });
    messagesBox.scrollTop = messagesBox.scrollHeight;
  });
}

function getMessages(func) {
  fetch(url + "/api/messages", {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
      // Add other headers as needed
    },
  }).then((response) => {
    response.json().then((data) => func(data));
  });
}

function sendMessage(msg, success) {
  fetch(url + "/api/messages", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      // Add other headers as needed
    },
    body: JSON.stringify(msg),
  }).then(() => {
    success();
  });
}
