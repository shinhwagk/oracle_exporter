const http = require("http");
const fs = require("fs");

const host = "dbmonitor.weihui.com";
const port = "3000";
const bak_dir = "grafana"
const authorization = { "Authorization": "Bearer eyJrIjoiNjBVV2V0Ukp0cUZ5QjN5MWoyZXFQZUhCOGVGZk5tcHAiLCJuIjoibG9hZCIsImlkIjoxfQ==" };

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
    const d = JSON.parse(data)
    if (!d.meta.isFolder) {
      for (let i = 0; i < d.dashboard.templating.list.length; i += 1) {
        d.dashboard.templating.list[0].current = { text: "", value: "" }
        d.dashboard.templating.list[0].options = []
      }
    }
    fs.writeFileSync(name, JSON.stringify(d), { encoding: "utf-8" });
  }
}

function storeDashboard(uri) {
  console.info(`/grafana/api/dashboards/${uri}`);
  httpClient(`/grafana/api/dashboards/${uri}`, storeD(`${bak_dir}\\${uri.split("/")[1]}.json`));
}

function storeDashboards(data) {
  for (let i of JSON.parse(data)) {
    // console.info("dashboard: ", i.uri);
    storeDashboard(i.uri)
  }
}

httpClient("/grafana/api/search", storeDashboards);