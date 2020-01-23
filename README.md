# ReconTools
Commandline tool for conducting passive recon against organizations. Recon data is collected using a numerous APIs, however it is also possible to run the tool without any keys specified (results will be limited).

```
git clone https://github.com/louisbarrett/recontools
cd recontools
go build 
./recontools

Usage of ./ReconTools:
  -banner
    	Show network banners
  -doxx
    	Attempt an OSINT look up on org CEO
  -employees
    	Attempt to discover employee profiles
  -network
    	Attempt to discover network perimiter via dig
  -org string
    	The name of the organization to scan
  -output string
    	Filename to output the report
  -ports
    	Attempt to discover network perimiter via dig
```

The following API keys should be set as environment variables to make use of advanced features
```
WIGLEAPIKEY - Wigle.net API KEY - https://api.wigle.net
WIGLEAPISECRET - Wigle.net API secret

CENSYSAPIKEY - Censys.io API KEY - https://censys.io/api
CENSYSAPISECRET - Censys.io API Secret

ABUSEDBSECRET - Abuse IP DB API Secret - https://www.abuseipdb.com/api

PTUSER - PassiveTotal (RiskIQ) API User - https://community.riskiq.com/
PTAPIKEY - PassiveTotal (RiskIQ) API Key

SHODANAPIKEY - Shodan API Key - https://shodan.io

HUNTERAPIKEY - Hunter.io API Key - https://hunter.io/api
```



```
recontools --org ServiceNow --network

Company Details

Name: ServiceNow, Inc.
CEO: John Donahoe
Founded: "2004"
Company TLD: servicenow.com
Industry Sector: null
Address: 2225 Lawson Lane  Santa Clara California

Network Perimeter from Dig
IP: 70.34.56.46 		 Hostname: ams20vcse01.servicenow.com
IP: 70.34.48.47 		 Hostname: sjc4vcse01.servicenow.com
IP: 10.195.224.154 		 Hostname: ams20uccorplabvcse01.servicenow.com
IP: 70.34.57.48 		 Hostname: lhr10ucexpye01.servicenow.com
IP: 70.34.49.48 		 Hostname: dal20ucexpye01.servicenow.com
IP: 70.34.60.61 		 Hostname: sin20ucexpye01.servicenow.com
IP: 70.34.53.48 		 Hostname: orl001ucexpye01.servicenow.com
IP: 70.34.56.38 		 Hostname: ams3ucexpye01.servicenow.com
IP: 70.34.48.48 		 Hostname: sjc4ucexpye01.servicenow.com
IP: 70.34.61.37 		 Hostname: syd4ucexpye01.servicenow.com
IP: 70.34.48.64 		 Hostname: sjc4seclogp01.servicenow.com
IP: 70.34.56.47 		 Hostname: ams20vcse02.servicenow.com
IP: 70.34.57.49 		 Hostname: lhr10ucexpye02.servicenow.com
IP: 70.34.49.49 		 Hostname: dal20ucexpye02.servicenow.com
...
```
