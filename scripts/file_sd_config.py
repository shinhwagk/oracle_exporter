import csv
import json
import cx_Oracle

import parms


class OracleDatabase:
    conn = None

    def __init__(self, ouser, opass, oip, oport, osvc):
        self.oip = oip
        self.oport = oport
        self.osvc = osvc
        self.ouser = ouser
        self.opass = opass
        self.createConnect()
        # self.closeConnect()

    def createConnect(self):
        self.conn = cx_Oracle.connect(
            "{}/{}@{}:{}/{}".format(self.ouser, self.opass, self.oip, self.oport, self.osvc))

    def closeConnect(self):
        self.conn.close()

    def getMeta(self):
        c = self.conn.cursor()
        c.execute("SELECT d.name, d.db_unique_name, d.database_role, i.version, i.instance_number, i.instance_name, i.host_name FROM v$instance i ,v$database d")
        name, uname, role, version, inst, inst_name, host = c.fetchone()
        c.close()
        return name.lower(), uname.lower(), role.lower(), version, str(inst), inst_name.lower(), host.lower()


def appendContainer(c, k, v):
    if c.get(k):
        c[k].append(v)
    else:
        c[k] = [v]


class OracleExporter:
    port_start_number = 9010
    ometa = None

    def __init__(self, ouser, opass, oip, oport, osvc, ozone, version, deployIp):
        self.oip = oip
        self.oport = oport
        self.osvc = osvc
        self.ozone = ozone
        self.version = version
        self.ouser = ouser
        self.opass = opass
        self.deployIp = deployIp
        self.ometa = OracleExporter.metaTemplate(
            *OracleDatabase(ouser, opass, oip, oport, osvc).getMeta())

    def generateCommands(self, container):
        node_exporeter_name = "{}_{}_{}".format(
            self.ometa['inst_name'], self.ometa['inst'], self.ometa['db_uname'])

        command = OracleExporter.commandTemplate(
            node_exporeter_name, self.port_start_number, self.ouser, self.opass, self.oip, self.oport, self.osvc, self.version)

        container.append(command)
        self.port_start_number += 1

    def generateFileSdConfig(self, container):
        target = "{}:{}".format(self.deployIp, self.port_start_number)
        oversion = self.ometa['version'].split('.')[0]
        isguard = self.ometa['db_role'] != 'PRIMARY'
        config = {
            "targets": [target],
            "labels": {"db_name": self.ometa['name'], "inst": self.ometa['inst'], "host": self.ometa['host'], 'vesrion': self.ometa['version'], 'db_uname': self.ometa['db_uname']}
        }

        groupName = 'oracle_'+oversion+'_dg' if isguard else oversion

        appendContainer(container, groupName, config)

        if isguard and self.ometa["inst"] == "1":
            groupName = 'oracle_'+oversion+'_dg_inst_1'

        appendContainer(container, groupName, config)

    @staticmethod
    def commandTemplate(uname, port, version, oip, oport, ouser, opass, osvc):
        return "./oracle_exporter.sh -n {} -p {} -c {}/{}@{}:{}/{} -v {}".format(uname, port, ouser, opass, oip, oport, svc, version)

    @staticmethod
    def metaTemplate(name, un_name, role, version, inst_num, inst_name, host):
        return {
            'name': name,
            'db_uname': un_name,
            'db_role': role,
            'version': version,
            "inst": inst_num,
            "inst_name": inst_name,
            "host": host
        }


def servers():
    with open('test.csv', encoding='UTF-8') as csvfile:
        spamreader = csv.reader(csvfile)
        return list(spamreader)


cc = {}
cm = []
for zone, ip, svc, b in servers():
    print(ip)
    if b == 'false':
        continue
    try:
        oe = OracleExporter(parms.username, parms.password,
                            ip, 1521, svc, zone, parms.version, parms.deployIp)
        oe.generateFileSdConfig(cc)
        # oe.generateCommands(cm)
        # od = OracleDatabase(ip, sve, zone)
        # od.createConnect()
        # od.closeConnect()
    except BaseException as e:
        print(e, "2222222222")
        print("{} {} connect exception".format(ip, svc))


print(json.dumps(cc))
print(json.dumps(cm))
# print(OracleExporter.start_scripts)
# query:
#   SELECT d.name, d.db_unique_name, d.database_role, i.version, i.instance_number, i.instance_name, i.host_name FROM v$instance i ,v$database d;


# def queryBaseInfo(ip):
#     con = cx_Oracle.connect()
#     cur = con.cursor()
#     return [""]


# def template(zone, name, un_name, role, version, inst_num, inst_name, host):
#     return {
#         'zone': zone,
#         'db_name': name,
#         'db_un_name': un_name,
#         'db_role': role,
#         'version': version,
#         "inst": inst_num,
#         "inst_name": inst_name,
#         "host": host
#     }
