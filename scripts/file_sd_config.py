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
        return name.lower(), uname.lower(), role.lower(), version[0:2], str(inst), inst_name.lower(), host.lower()


def appendContainer(c, k, v):
    if c.get(k):
        c[k].append(v)
    else:
        c[k] = [v]


class OracleExporter:

    ometa = None

    def __init__(self, ouser, opass, oip, oport, osvc, ozone, version, deployIp, deployPort):
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
        self.deployPort = deployPort

    def generateCommands(self, container):
        node_exporeter_name = "{}_{}_{}".format(
            self.ometa['name'], self.ometa['inst'], self.ometa['db_uname'])

        command = OracleExporter.commandTemplate(
            node_exporeter_name, self.deployPort, self.version, self.oip, self.oport, self.osvc, self.ouser, self.opass)

        container.append(command)

    def generateFileSdConfig(self, container):
        target = "{}:{}".format(self.deployIp, self.deployPort)
        oversion = self.ometa['version'].split('.')[0]
        dbrole = ''.join([i[0] for i in self.ometa['db_role'].split(' ')])
        config = {
            "targets": [target],
            "labels": {"name": self.ometa['name'], "inst": self.ometa['inst'], 'vesrion': self.ometa['version'], 'role': dbrole}
        }

        if self.ometa["db_role"] == "primary":
            groupName = 'oracle_'+oversion+"_p"
            appendContainer(container, groupName, config)

        if self.ometa['inst'] == '1' and self.ometa["db_role"] == "primary":
            groupName = 'oracle_'+oversion+"_p_i1"
            appendContainer(container, groupName, config)

        if self.ometa["db_role"] == "physical standby":
            groupName = 'oracle_'+oversion+'_dg_ps'
            appendContainer(container, groupName, config)

        if self.ometa["db_role"] == "physical standby" and self.ometa["inst"] == "1":
            groupName = 'oracle_'+oversion+'_dg_ps_i1'
            appendContainer(container, groupName, config)

        if self.ometa["db_role"] == "logical standby":
            groupName = 'oracle_' + oversion + '_dg_ls'
            appendContainer(container, groupName, config)

    @staticmethod
    def commandTemplate(uname, port, version, oip, oport, osvc, ouser, opass):
        return "./oracle_exporter.sh -n {} -p {} -c {}/{}@{}:{}/{} -v {}".format(uname, port, ouser, opass, oip, oport, osvc, version)

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
    with open('servers.csv', encoding='UTF-8') as csvfile:
        spamreader = csv.reader(csvfile)
        return list(spamreader)


def main():
    cc = {}
    cm = []
    port_start_number = parms.oexporter_start_post
    for zone, ip, svc, b in servers():
        print(ip)
        if b == 'false':
            continue
        try:
            oe = OracleExporter(parms.username, parms.password,
                                ip, 1521, svc, zone, parms.version, parms.deployIp, port_start_number)
            oe.generateFileSdConfig(cc)
            oe.generateCommands(cm)
            port_start_number += 1
            # od = OracleDatabase(ip, sve, zone)
            # od.createConnect()
            # od.closeConnect()
        except BaseException as e:
            print("{} {} connect exception: {}".format(ip, svc, e))

    print(json.dumps(cc))
    for m in cm:
        print(m)
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


if __name__ == "__main__":

    main()
