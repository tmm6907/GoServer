# Open-NWI API

## Purporse
An open-source API to access EPA's National Walkability Index for address in the US.
## Usage
### Input
```
//Search by address : Sreet, City, Zip
curl -o out.json "localhost:8080/scores/address?address=1600%20Pennsylvania%20Avenue%20NW%2C%20Washington%2C%20DC%2020500%20"

//Search by address : Sreet, City
curl -o out.json "localhost:8080/scores/address?address=1600%20Pennsylvania%20Avenue%20NW%2C%20Washington%2C%20DC"

//Get list of scores [limit 50 by default]
curl -o out.json "localhost:8080/scores"

//Get list of scores with limit and offset
curl -o out.json "localhost:8080/scores?limit=3&offset=13000"
```

### Output *out.json*
Where **nwi** is the total score of walkability for the given area.
```json
{
  "geoid": 110010062021,
  "nwi": 18,
  "searchedAddress": "1600 Pennsylvania Avenue NW, Washington, DC 20500"
}

{
  "geoid": 110010062021,
  "nwi": 18,
  "searchedAddress": "1600 Pennsylvania Avenue NW, Washington DC"
}

[
  {
    "id": 13000,
    "geoid": 484599504002,
    "csa_name": "",
    "cbsa_name": "Longview, TX",
    "nwi": 6.166666666666667
  },
  {
    "id": 13001,
    "geoid": 484599505001,
    "csa_name": "",
    "cbsa_name": "Longview, TX",
    "nwi": 3.5
  },
  {
    "id": 13002,
    "geoid": 484599501006,
    "csa_name": "",
    "cbsa_name": "Longview, TX",
    "nwi": 2.8333333333333335
  }
]
```
