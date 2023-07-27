# Open-NWI API

## Purporse
An open-source API to access EPA's National Walkability Index for any address recognized by US Census Geocoding. Checkout https://opennwi.dev/ to get started!
## Usage
### Input
/scores/?address=1600 Pennsylvania Avenue NW, Washington, DC
/scores/?address=1600 Pennsylvania Avenue NW, Washington, DC&format=xml
/scores/?zipcode=20050&limit=3&format=json
/scores/?zipcode=20050&limit=3&format=json&offset=130

### Output *out.json*
Where **nwi** is the total score of walkability for the given area. Scores range from 0-20.
```js
// address with zipcode
{
  rankID:61619
  geoid:110010062021
  cbsaName:"Washington-Arlington-Alexandria, DC-VA-MD-WV"
  nwi:18
  transitScore:20
  bikeScore:9
  searchedAddress:"1600 Pennsylvania Avenue NW, Washington, DC 20500"
  format:"json"
}

// address without zipcode (format=xml)
<result>
  <rankID>61619</rankID>
  <geoid>110010062021</geoid>
  <cbsaName>Washington-Arlington-Alexandria, DC-VA-MD-WV</cbsaName>
  <nwi>18</nwi>
  <transitScore>20</transitScore>
  <bikeScore>9</bikeScore>
  <searchedAddress>1600 Pennsylvania Avenue NW, Washington, DC</searchedAddress>
  <format>xml</format>
</result>

// list of scores (limit=3, format=xml)
<results>
  <result id="0">
    <rankID>115880</rankID>
    <geoid>271630710031</geoid>
    <cbsaName>Minneapolis-St. Paul-Bloomington, MN-WI</cbsaName>
    <nwi>8.5</nwi>
    <transitScore>19</transitScore>
    <bikeScore>17</bikeScore>
    <format>xml</format>
  </result>
  <result id="1">
    <rankID>115881</rankID>
    <geoid>271630710032</geoid>
    <cbsaName>Minneapolis-St. Paul-Bloomington, MN-WI</cbsaName>
    <nwi>8.333333333</nwi>
    <transitScore>19</transitScore>
    <bikeScore>17</bikeScore>
    <format>xml</format>
  </result>
  <result id="2">
    <rankID>115882</rankID>
    <geoid>271630710033</geoid>
    <cbsaName>Minneapolis-St. Paul-Bloomington, MN-WI</cbsaName>
    <nwi>6.833333333</nwi>
    <transitScore>19</transitScore>
    <bikeScore>17</bikeScore>
    <format>xml</format>
  </result>
</results>

// list of scores (limit=3, offset=130, format=json)
[
  {
    id:130
    rankID:15244
    geoid:511076118032
    cbsaName:"Washington-Arlington-Alexandria, DC-VA-MD-WV"
    nwi:11.33333333
    transitScore:20
    bikeScore:9
    format:"json"
  }
  {
    id:131
    rankID:15258
    geoid:511790102053
    cbsaName:"Washington-Arlington-Alexandria, DC-VA-MD-WV"
    nwi:7.5
    transitScore:20
    bikeScore:9
    format:"json"
  }
  {
    id:132
    rankID:15259
    geoid:511790102132
    cbsaName:"Washington-Arlington-Alexandria, DC-VA-MD-WV"
    nwi:5.166666667
    transitScore:20
    bikeScore:9
    format:"json"
  }
]
