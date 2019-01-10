local hostname = "http://star-otepon00.qe.ogk.ynwm.yahoo.co.jp";
local http_port = 8080;
local jvm_xmx = "16G";

{
    local ConfigBase = {
        "http-server.http.port": http_port,
        "discovery.uri": hostname + ":" + http_port,
    },
    // config.properties
    "coordinator/config.properties":
        std.manifestIni({
          main: ConfigBase + {
            "coordinator":"true",
            "node-scheduler.include-coordinator":"false",
            "discovery-server.enabled":"true",
          },
          sections:{}
        })
    ,
    "worker/config.properties":
            std.manifestIni({
              main: ConfigBase + {
                "coordinator":"false",
              },
              sections:{}
            })
        ,

    // jvm.config
    "coordinator/jvm.config": |||
      -server
      -Xmx%(xmx)s
      -XX:-UseBiasedLocking
      -XX:+UseG1GC
      -XX:+ExplicitGCInvokesConcurrent
      -XX:+UseGCOverheadLimit
      -XX:+ExitOnOutOfMemoryError
      -XX:ReservedCodeCacheSize=512M
    ||| % {xmx:jvm_xmx},
    "worker/jvm.config": $["coordinator/jvm.config"],

    // node.properties
    "coordinator/node.properties":
       std.manifestIni({
          main: {
            "node.environment":"test",
            "node.id":"90508676-d288-4782-a5b6-405a5f6d2fd6",
            "node.data-dir":"/var/lib/presto/data",
            "catalog.config-dir":"/etc/presto/catalog",
            "plugin.dir":"/usr/lib/presto/lib/plugin",
            "node.server-log-file":"/var/log/presto/server.log",
            "node.launcher-log-file":"/var/log/presto/launcher.log",
          },
          sections:{}
        }),
    "worker/node.properties": $["coordinator/node.properties"],
}