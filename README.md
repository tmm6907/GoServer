# Open-NWI API

## Purporse
An open-source API to access EPA's National Walkability Index for any address recognized by US Census Geocoding.
## Usage
### Input
```
// Search by address : Sreet, City, Zip
curl -o out.json "localhost:8080/scores/address?address=1600%20Pennsylvania%20Avenue%20NW%2C%20Washington%2C%20DC%2020500%20"

// Search by address : Sreet, City
curl -o out.json "localhost:8080/scores/address?address=1600%20Pennsylvania%20Avenue%20NW%2C%20Washington%2C%20DC"

// Get list of scores (limit=50)
curl -o out.json "localhost:8080/scores"

// Get list of scores (limit=3, offset=13000)
curl -o out.json "localhost:8080/scores?limit=3&offset=13000"
```

### Output *out.json*
Where **nwi** is the total score of walkability for the given area.
```js
// address with zipcode
{
  "geoid": 110010062021,
  "nwi": 18,
  "searchedAddress": "1600 Pennsylvania Avenue NW, Washington, DC 20500"
}

// address without zipcode
{
  "geoid": 110010062021,
  "nwi": 18,
  "searchedAddress": "1600 Pennsylvania Avenue NW, Washington DC"
}

// list of scores (limit=50)
[
  {
    "id": 0,
    "geoid": 481130078254,
    "csa_name": "Dallas-Fort Worth, TX-OK",
    "cbsa_name": "Dallas-Fort Worth-Arlington, TX",
    "nwi": 14
  },
  {
    "id": 1,
    "geoid": 481130078252,
    "csa_name": "Dallas-Fort Worth, TX-OK",
    "cbsa_name": "Dallas-Fort Worth-Arlington, TX",
    "nwi": 10.833333333333334
  },
  {
    "id": 2,
    "geoid": 481130078253,
    "csa_name": "Dallas-Fort Worth, TX-OK",
    "cbsa_name": "Dallas-Fort Worth-Arlington, TX",
    "nwi": 8.333333333333334
  },
  {
    "id": 3,
    "geoid": 481130078241,
    "csa_name": "Dallas-Fort Worth, TX-OK",
    "cbsa_name": "Dallas-Fort Worth-Arlington, TX",
    "nwi": 15.666666666666668
  },
  {
    "id": 4,
    "geoid": 481130078242,
    "csa_name": "Dallas-Fort Worth, TX-OK",
    "cbsa_name": "Dallas-Fort Worth-Arlington, TX",
    "nwi": 10.166666666666668
  },
  {
    "id": 5,
    "geoid": 481130078271,
    "csa_name": "Dallas-Fort Worth, TX-OK",
    "cbsa_name": "Dallas-Fort Worth-Arlington, TX",
    "nwi": 6.833333333333333
  },
  ...
]

// list of scores (limit=3, offset=13000)
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
