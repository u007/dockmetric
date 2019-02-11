#!/bin/bash
export USER=admin
export PASSWORD=$GRAFANA_PASS
export HOST=127.0.0.1:3000

# nc -z 127.0.0.1 3000 &> /dev/null
curl -s $HOST > /dev/null
while [ $? -ne 0 ]; do
  echo "waiting..."
  sleep 1
  curl -s $HOST > /dev/null
done

set -e
echo "changing admin password"
curl -X PUT -H "Content-Type: application/json" -d "{
  \"oldPassword\": \"admin\",
  \"newPassword\": \"$PASSWORD\",
  \"confirmNew\": \"$PASSWORD\"
}" admin:admin@$HOST/api/user/password

echo "adding user mradmin"
curl -X POST -H "Content-Type: application/json" -d "{
  \"name\": \"mradmin\",
  \"login\": \"mradmin\",
  \"password\": \"$PASSWORD\",
  \"isAdmin\": false,
  \"orgId\": 1
}" $USER:$PASSWORD@$HOST/api/admin/users


# setup organisation
# curl -X POST -H "Content-Type: application/json" -d '{"name":"default"}' $HOST/api/orgs

# setup influxdb
echo "adding plugins..."
curl -d '{"orgId":1,"name":"InfluxDB","type":"influxdb","typeLogoUrl":"public/app/plugins/datasource/influxdb/img/influxdb_logo.svg","access":"proxy","url":"http://influx:8086","password":"007890","user":"dbu1","database":"report","basicAuth":false,"isDefault":true,"jsonData":{"keepCookies":[]},"readOnly":false}' -H "Content-Type: application/json" -X POST $USER:$PASSWORD@$HOST/api/datasources


echo "setting up dashboard..."

PANELJSON='{  
  "aliasColors":{  

  },
  "id": $id,
  "bars":false,
  "dashLength":10,
  "dashes":false,
  "datasource": "InfluxDB",
  "fill":1,
  "gridPos":{  
    "h":14,
    "w":23,
    "x":0,
    "y":0
  },
  "legend":{  
    "avg":false,
    "current":false,
    "max":false,
    "min":false,
    "show":true,
    "total":false,
    "values":false
  },
  "lines":true,
  "linewidth":1,
  "links":[

  ],
  "nullPointMode":"null",
  "percentage":false,
  "pointradius":5,
  "points":false,
  "renderer":"flot",
  "seriesOverrides":[  

  ],
  "spaceLength":10,
  "stack":false,
  "steppedLine":false,
  "targets":[  
    {  
      "groupBy":[  
        {  
          "params":[  
            "$__interval"
          ],
          "type":"time"
        },
        {  
          "params":[  
            "null"
          ],
          "type":"fill"
        }
      ],
      "measurement":"containers",
      "orderByTime":"ASC",
      "policy":"6months",
      "refId":"A",
      "resultFormat":"time_series",
      "select":[  
        [  
          {  
            "params":[  
              "cpu_percent"
            ],
            "type":"field"
          },
          {  
            "params":[  

            ],
            "type":"mean"
          },
          {  
            "params":[  
              "cpu"
            ],
            "type":"alias"
          }
        ],
        [  
          {  
            "params":[  
              "io_read"
            ],
            "type":"field"
          },
          {  
            "params":[  

            ],
            "type":"mean"
          },
          {  
            "params":[  
              "disk read"
            ],
            "type":"alias"
          }
        ],
        [  
          {  
            "params":[  
              "io_write"
            ],
            "type":"field"
          },
          {  
            "params":[  

            ],
            "type":"mean"
          },
          {  
            "params":[  
              "disk write"
            ],
            "type":"alias"
          }
        ],
        [  
          {  
            "params":[  
              "memory"
            ],
            "type":"field"
          },
          {  
            "params":[  

            ],
            "type":"mean"
          },
          {  
            "params":[  
              "memory"
            ],
            "type":"alias"
          }
        ],
        [  
          {  
            "params":[  
              "net_read"
            ],
            "type":"field"
          },
          {  
            "params":[  

            ],
            "type":"mean"
          },
          {  
            "params":[  
              "network in"
            ],
            "type":"alias"
          }
        ],
        [  
          {  
            "params":[  
              "net_write"
            ],
            "type":"field"
          },
          {  
            "params":[  

            ],
            "type":"mean"
          },
          {  
            "params":[  
              "network out"
            ],
            "type":"alias"
          }
        ]
      ],
      "tags": [
        {
          "key": "name",
          "operator": "=",
          "value": "$containerName"
        }
      ]
    }
  ],
  "thresholds":[  

  ],
  "timeFrom":null,
  "timeRegions":[  

  ],
  "timeShift":null,
  "title":"$panelname",
  "tooltip":{  
    "shared":true,
    "sort":0,
    "value_type":"individual"
  },
  "type":"graph",
  "xaxis":{  
    "buckets":null,
    "mode":"time",
    "name":null,
    "show":true,
    "values":[  

    ]
  },
  "yaxes":[  
    {  
      "format":"short",
      "label":null,
      "logBase":1,
      "max":null,
      "min":null,
      "show":true
    },
    {  
      "format":"short",
      "label":null,
      "logBase":1,
      "max":null,
      "min":null,
      "show":true
    }
  ],
  "yaxis":{  
    "align":false,
    "alignLevel":null
  }
}'


echo "containers: $CONTAINERS"
if [ "$CONTAINERS" == "" ]; then
  echo "Please declare CONTAINERS separated by comma"
  exit 1
fi

containerRes=""
let "containerNo=1"
for i in $(echo $CONTAINERS | sed "s/,/ /g")
do

  if [ "$containerRes" != "" ]; then
    containerRes=$containerRes",
"
  fi
  panelRes="${PANELJSON/\$panelname/$i}"
  panelRes="${panelRes/\$containerName/$i}"
  panelRes="${panelRes/\$id/$containerNo}"
  
  containerRes=$containerRes$panelRes
  let "containerNo=containerNo+1"
done

dashboardRes=`cat /dashboard.raw.json`
dashboardRes="${dashboardRes/\$panels/$containerRes}"

echo $dashboardRes > /dashboard.json
cat /dashboard.json

export dashboardJSON=$(curl -X POST -H "Content-Type: application/json" -d @/dashboard.json $USER:$PASSWORD@$HOST/api/dashboards/db)

echo "dashboard: $dashboardJSON"
export dashboardID=`echo $dashboardJSON | jq '.id'`
# {"id":1,"slug":"containers-overview","status":"success","uid":"nZiODFXik","url":"/d/nZiODFXik/containers-overview","version":1}

echo "dashboard-id: $dashboardID"
curl -X PUT /api/org/preferences -H "Content-Type: application/json" -d "{
  \"theme\": \"\",
  \"homeDashboardId\": $dashboardID,
  \"timezone\":\"Singapore\"
}" $USER:$PASSWORD@$HOST/api/org/preferences

touch /custom/installed

echo "installed!"

set +e
