function postData(path, name, value) {
  let data = { long_url: document.getElementById("input").value, user_id: "0" };

  fetch("http://localhost:9808/create-short-url", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  })
    .then((res) => res.json())
    .then((data) => {
      console.log(data.short_url);
      var output = document.getElementById("output");
      output.href = data.short_url;
      output.text = data.short_url;
      output.style.display = "block";
    })
    .catch((err) => {
      console.log(err);
    });
}
