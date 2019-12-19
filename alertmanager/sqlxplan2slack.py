from http.server import HTTPServer, BaseHTTPRequestHandler
import json
import os
import time
import datetime


from slackclient import SlackClient
import requests


host = ('localhost', 8888)

XPLAN_URL = "http://127.0.0.1:3000/xplan_text"
SLACK_API_TOKEN = os.environ["SLACK_API_TOKEN"]
SLACK_INCOMING_WEBHOOK_URL = ""
# SLACK_WORKSPACE_NAME = "dba"
CHANNEL_ASHSQLFILE = ""
GRAFANA_URL = "http://xxx:3000/grafana/d/r1_mzGHiz/oracle-sql-detail"


sc = SlackClient(SLACK_API_TOKEN)


class WebHook(BaseHTTPRequestHandler):
    def do_POST(self):
        data = self.rfile.read(int(self.headers['content-length'])).decode('utf-8')
        sendAlertToSlack(json.loads(data))
        self.send_response(200)
        self.end_headers()
        self.wfile.write('ok'.encode('UTF-8'))

    def do_GET(self):
        self.send_response(201)
        self.end_headers()
        self.wfile.write('ok'.encode('UTF-8'))


def alertAction_PRTM(alertSchema):
    name = alertSchema["labels"]["alertname"]
    db_name = alertSchema["labels"]["db_name"]
    inst = alertSchema["labels"]["inst"]
    username = alertSchema["labels"]["username"]
    opname = alertSchema["labels"]["opname"]
    value = alertSchema["annotations"]["value"]
    status = alertSchema["status"]
    sql_id = alertSchema["labels"]["sql_id"]
    child = alertSchema["labels"]["child"]
    description = alertSchema["annotations"]["description"]
    title_link = sqlDetailUrl(db_name, inst, username, sql_id, child)
    xplanText = getXPlanText(db_name, inst, sql_id, child)
    uploadFileName = "{0}-{1}-{2}-{3}".format(db_name, inst, sql_id, child)
    url_private = uploadFileCotnentToSlack(uploadFileName, xplanText)
    x = "{0} - {1} - {2} - {3} - {4}".format(db_name, inst, username, opname, sql_id)
    slackMessage = slackTemplate(name, status, title_link, description, value, url_private, x)
    slackIncommingHook(slackMessage)
    print("send success: %s" % uploadFileName)


def alertAction_OASUS(alertSchema):
    name = alertSchema["labels"]["alertname"]
    status = alertSchema["status"]
    db_name = alertSchema["labels"]["db_name"]
    inst = alertSchema["labels"]["inst"]
    username = alertSchema["labels"]["username"]
    opname = alertSchema["labels"]["opname"]
    value = alertSchema["annotations"]["value"]
    status = alertSchema["status"]
    sql_id = alertSchema["labels"]["sql_id"]
    event = alertSchema["labels"]["event"]
    description = alertSchema["annotations"]["description"]
    title_link = sqlDetailUrl(db_name, inst, username, sql_id)
    xplanText = getXPlanText(db_name, inst, sql_id)
    uploadFileName = "{0}-{1}-{2}-{3}".format(db_name, inst, sql_id, 0)
    url_private = uploadFileCotnentToSlack(uploadFileName, xplanText)
    x = "{0} - {1} - {2} - {3} - {4} - {5}".format(db_name, inst, username, opname, sql_id, event)
    slackMessage = slackTemplate(name, status, title_link, description, value, url_private, x)
    slackIncommingHook(slackMessage)
    print("send success: %s" % uploadFileName)


alertActions = {
    "Oracle Active Session User SQL": alertAction_OASUS,
    "Physical Read Too Many": alertAction_PRTM}


def sendAlertToSlack(alertmangerMessage):
    for alertSchema in alertmangerMessage["alerts"]:
        alertname = alertSchema["labels"]["alertname"]
        if alertname in alertActions:
            action = alertActions.get(alertname)
            action(alertSchema)


def uploadFileCotnentToSlack(name, content):
    """
    api: https://api.slack.com/methods/files.upload
    """
    res = sc.api_call("files.upload", channels=CHANNEL_ASHSQLFILE, content=content, title=name, filetype="text")
    return res["file"]["url_private"]


def slackIncommingHook(incomingWebhookSendArguments):
    requests.post(SLACK_INCOMING_WEBHOOK_URL, data=json.dumps(incomingWebhookSendArguments))


def slackTemplate(name, status, title_link, description, value, url_private, x):
    return {
        "attachments": [{
            "title": "[{0}] {1}".format(status, name),
            "title_link": title_link,  # ,
            "text": description,
            "color": "#D63232" if status == "firing" else "#36a64f",
            "fields": [
                {"title": x, "value": value, "short": True}
            ],
            "actions": [{
                "text": "SQL_PLAN",
                "type": "button",
                "url": url_private
            }],
            "footer":      "AlertManager v1.0.0",
            "footer_icon": "https://a.slack-edge.com/7f1a0/plugins/app/assets/service_32.png",
            "ts": time.mktime(datetime.datetime.now().timetuple())
        }]
    }


def sqlDetailUrl(db_name, inst, username, sql_id, child=0):
    # payload = {"var-db_name": db_name, "var-inst": inst, "var-sql_id": sql_id, "var-child": 0}
    url = "{0}?var-db_name={1}&var-inst={2}&var-username={3}&var-sql_id={4}&var-child={5}".format(GRAFANA_URL, db_name, inst, username, sql_id, child)
    return url


def getXPlanText(db_name, inst, sql_id, child=0):
    payload = {"db_name": db_name, "inst": inst, "sql_id": sql_id, "child": child}
    r = requests.get(XPLAN_URL, params=payload)
    return r.text


if __name__ == '__main__':
    # iid = uploadFileCotnentToSlack("xx", "zfjs-1-fdwm60s7n4pyv-1")
    # slackIncommingHook({"text": "xxxx"})
    server = HTTPServer(host, WebHook)
    print('Prometheus Alertmanger Webhook For <Oracle ash-sql>, listen at: %s:%s' % host)
    server.serve_forever()
