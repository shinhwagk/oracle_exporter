const http = require("http");
const fs = require("fs");

const host = "10.65.193.52";
const port = "3000";
const bak_dir = "grafana"
const authorization = { "Authorization": "Bearer eyJrIjoiYkdiQnBhbklFdFV3TUhLdjI0WUJRUVIyUDBXVWhybzIiLCJuIjoiYWRtaW4iLCJpZCI6MX0=" };

function httpClient(url, func) {
  http.get(
    {
      host: host,
      port: port,
      path: url,
      headers: authorization
    }, (res) => {
      let data = "";
      res.on("data", (chunk) => data += chunk)
      res.on('end', () => func(data));
      res.on("error", (err) => console.info(err))
    })
};

function storeD(name) {
  return (data) => {
    fs.writeFileSync(name, JSON.stringify(JSON.parse(data).dashboard), { encoding: "utf-8" });
  }
}

function store(uri) {
  console.info(`/api/dashboards/${uri}`);
  httpClient(`/api/dashboards/${uri}`, storeD(`${bak_dir}\\${uri.split("/")[1]}.json`));
}

function getD(data) {
  for (let i of JSON.parse(data)) {
    console.info("d", i.uri);
    store(i.uri)
  }
}

httpClient("/api/search", getD);