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

    async def closeConnect(self):
        self.conn.close()

    def getMeta(self):
        c = self.conn.cursor()
        c.execute("SELECT d.name, d.db_unique_name, d.database_role, i.version, i.instance_number, i.instance_name, i.host_name FROM v$instance i ,v$database d")
        name, uname, role, version, inst, inst_name, host = c.fetchone()
        c.close()
        self.closeConnect()
        return {"name": name.lower(), "uname": uname.lower(), "role": role.lower(), "version": version[0:2], "inst": str(inst), "inst_name": inst_name.lower(), "host": host.lower()}


def appendContainer(c, k, v):
    if c.get(k):
        c[k].append(v)
    else:
        c[k] = [v]


class OracleExporter:

    def __init__(self, ouser, opass, ogroup, oip, oport, osvc, ozone, deployVersion, deployIp, deployPort):
        self.oip = oip
        self.oport = oport
        self.ogroup = ogroup
        self.osvc = osvc
        self.ozone = ozone
        self.deployVersion = deployVersion
        self.ouser = ouser
        self.opass = opass
        self.deployIp = deployIp
        self.ometa = OracleDatabase(ouser, opass, oip, oport, osvc).getMeta()

        self.deployPort = deployPort

    # def generateCommands(self, container):
    #     node_exporeter_name = "{}_{}_{}".format(
    #         self.ometa['name'], self.ometa['inst'], self.ometa['db_uname'])

    #     command = OracleExporter.commandTemplate(
    #         node_exporeter_name, self.deployPort, self.version, self.oip, self.oport, self.osvc, self.ouser, self.opass)

    #     container.append(command)

    def generateFileSdConfig(self, container):
        target = "{}:{}".format(self.deployIp, self.deployPort)
        dbrole = ''.join([i[0] for i in self.ometa['db_role'].split(' ')])
        config = {
            "targets": [target],
            "labels": {"db_uname": self.ometa['uname'], "db_inst": self.ometa['inst'], 'db_vesrion': self.ometa['version'], 'db_role': dbrole, "db_group": self.ogroup}
        }

        oversion = self.ometa['version']
        if oversion in ["10", "11"]:
            oversion += 'g'

        if int(oversion) >= 12:
            oversion += 'c'

        groupName = 'oracle_'+oversion
        appendContainer(container, groupName, config)

        if self.ometa["db_role"] == "primary":
            groupName = 'oracle_'+oversion+"_p"
            appendContainer(container, groupName, config)

            if self.ometa['inst'] == '1':
                groupName = 'oracle_'+oversion+"_p_i1"
                appendContainer(container, groupName, config)

        if self.ometa["db_role"] == "physical standby":
            groupName = 'oracle_'+oversion+'_dg_ps'
            appendContainer(container, groupName, config)

            if self.ometa["inst"] == "1":
                groupName = 'oracle_'+oversion+'_dg_ps_i1'
                appendContainer(container, groupName, config)

        if self.ometa["db_role"] == "logical standby":
            groupName = 'oracle_' + oversion + '_dg_ls'
            appendContainer(container, groupName, config)

    # def commandTemplate(self, uname, port, version, oip, oport, osvc, ouser, opass):
    #     return "./oracle_exporter.sh -n {} -p {} -c {}/{}@{}:{}/{} -v {}".format(uname, port, ouser, opass, oip, oport, osvc, version)

    def metaTemplate(self, ogroup, o_uname, role, version, inst_num):
        return {
            'db_group': ogroup,
            'db_uname': o_uname,
            'db_role': role,
            'db_version': version,
            "db_inst": inst_num
        }


def servers():
    with open('servers.csv', encoding='UTF-8') as csvfile:
        spamreader = csv.reader(csvfile)
        return list(spamreader)


def main():
    file_configs = {}
    # oracle_exporter_commands = []
    port_start_number = parms.oexporter_start_post
    for o_zone, o_ip, o_service, is_configed, o_group in servers():
        print(o_ip)
        if is_configed == 'false':
            continue
        try:
            oe = OracleExporter(parms.username, parms.password, o_group,
                                o_ip, 1521, o_service, o_zone, parms.version, parms.deployIp, port_start_number)
            oe.generateFileSdConfig(file_configs)
            # oe.generateCommands(oracle_exporter_commands)
            port_start_number += 1
        except BaseException as e:
            print("{} {} connect exception: {}".format(o_ip, o_service, e))

    for name, config in file_configs.items():
        f = open(name, '+w')
        f.write(json.dumps(config))
        f.close()


if __name__ == "__main__":
    main()
