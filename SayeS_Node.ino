/*
    HTTP over TLS (HTTPS) example sketch

    This example demonstrates how to use
    WiFiClientSecure class to access HTTPS API.
    We fetch and display the status of
    esp8266/Arduino project continuous integration
    build.

    Created by Ivan Grokhotkov, 2015.
    This example is in public domain.
*/

#include <ESP8266WiFi.h>
#include <WiFiClientSecure.h>
#include <SoftwareSerial.h>

const char* ssid = "Seif";
const char* password = "01111227871";

const char* host = "5ee5c7ca.ngrok.io";
const int httpsPort = 443;
String url = "/updateSpots";

String is_pending_str = "is_pending" ; 
String pending_spots_str = "pending_spots";
String sendUpdates(String data) {
  // Use WiFiClientSecure class to create TLS connection
  WiFiClientSecure client;
  //  Serial.print("connecting to ");
  //  Serial.println(host);
  if (!client.connect(host, httpsPort)) {
    //    Serial.println("connection failed");
    return "";
  }
  //  Serial.print("requesting URL: ");
  //  Serial.println(url);
  client.print(String("POST ") + url + " HTTP/1.1\r\n" +
               "Host: " + host + "\r\n" +
               "User-Agent: BuildFailureDetectorESP8266\r\n" +
               "Connection: close\r\n" +
               "Content-Type: text/html\r\n" +
               "Content-Length: " + data.length() + "\r\n" +
               "\r\n" +
               data + "\n");
  //  Serial.println("request sent");
  while (client.connected()) {
    String line = client.readStringUntil('\n');
    if (line == "\r") {
      //      Serial.println("headers received");
      break;
    }
  }
  String line = client.readStringUntil('\n');
  //  Serial.println("reply was:");
  //  Serial.println("==========");
  return line ;
  //  Serial.println("==========");
  //  Serial.println("closing connection");
}
void WifiSetup() {
  //  Serial.println();
  //  Serial.print("connecting to ");
  //  Serial.println(ssid);
  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    //    Serial.print(".");
  }
   digitalWrite(D0, HIGH);
  //  Serial.println("");
  //  Serial.println("WiFi connected");
  //  Serial.println("IP address: ");
  //  Serial.println(WiFi.localIP());
}
String getBoolean (String s ) {
  //14 - 19

    return s.substring(14, 19);
  
}


String getPending(String s) {
  String res = "" ;
  for (int index = 36 ; index < s.length() && s[index] != '\"'; index ++ ) {
    res += s[index];
  }
  return res;
}

void setup() {
  Serial.begin(9600);
  pinMode(D0, OUTPUT);
  WifiSetup();
  while (!Serial) {
    ; // wait for serial port to connect. Needed for native USB port only
  }
}
bool startreading = false ;
String arduino_spots = "";
char c;
String pending_spots;
String parsedString;
int i;
String flag ;
void startNode () {
  if (Serial.available()) {
    c = Serial.read();
    if (startreading)
      arduino_spots += c;
    if (c == '<')
      startreading = true ;
    else if (c == '>') {
      startreading = false ;
      pending_spots = sendUpdates(arduino_spots.substring(0, arduino_spots.length() - 1));
      flag = getBoolean(pending_spots);
      if (flag.equals("true,")||flag.equals("true")) {
        parsedString = getPending(pending_spots);
        for ( i = 0; i < parsedString.length(); i++)
          Serial.write(parsedString.charAt(i));
      }
      else {
        for ( i = 0; i < flag.length(); i++)
          Serial.write(flag.charAt(i));
      }
      arduino_spots = "";
    }
  }
}
//String error = "error";
//String parsed ;
//int myindex ;
//int secIndex;
//int startIndex;
//int endIndex;
//String temp;
//String parseString (String attr,String Json, bool flag){
////if (attr.length() > Json.length() ) 
////return error;   
//
//for (myindex = 0 ; myindex < Json.length() - attr.length(); myindex ++ ){
//  temp = Json.substring(myindex,myindex+attr.length()-1);
//  if (temp.equals(attr)){
//    // return value
//    startIndex = -1; 
//    endIndex = -1 ;
//    for (secIndex = myindex+attr.length() ;secIndex<Json.length();secIndex++){
//      if (startIndex>-1){
//        if (flag)
//        if (Json[secIndex] == '\"'){
//          endIndex = secIndex; 
//          break;
//        } else if (Json[secIndex] == ','){
//           endIndex = secIndex; 
//          break;
//        }
//        
//      }
//      else {
//        if(isAlphaNumeric(Json[secIndex])){
//          startIndex = secIndex;
//        }
//      }
//    }
//    return Json.substring(startIndex,endIndex);
//  }
//}
//return error;
//}
void loop() {
  startNode();
}
