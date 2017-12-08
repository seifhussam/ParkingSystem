
#include <SoftwareSerial.h>

class Color {
  public : bool red, green, blue;
    Color (bool r, bool g, bool b) {
      red =  r;
      green =  g;
      blue =  b;
    }
};
Color color_red (true, false, false);
Color color_green (false, true, false);
Color color_blue(false, false, true);
Color color_white(false, false, false);


/*
   status 0 = available    // blinking green
   status 1 = not availble // blinking red
   status 2 = pending   // blinking yellow
*/
int one_minute = 60000;
int delay_led = 200;
int minute = 240;
class RGBled
{
  public:
    int Red;
    int Green;
    int Blue;
    int redPin;
    int greenPin;
    int bluePin;
    int led_status;
    int id;
    int sensorPin;
    int c ;
    RGBled(int id1, int redPin1, int greenPin1, int bluePin1, int sensorPin1)
    {
      id = id1;
      Red = color_green.red;
      Green = color_green.green;
      Blue = color_green.blue;
      led_status = 0;
      c = 0;
      redPin = redPin1;
      greenPin = greenPin1;
      bluePin = bluePin1;
      sensorPin = sensorPin1;
      pinMode(sensorPin, INPUT);
      pinMode(redPin, OUTPUT);
      pinMode(greenPin, OUTPUT);
      pinMode(bluePin, OUTPUT);
    }
    void start() {
      
      turnOn();
    }
    void turnOn() {
      
      if (!digitalRead(sensorPin)) {
        setOccupied();
      } else {
        switch (led_status) {
          case 0 :setAvailable() ; break;
          case 1 : setAvailable() ; break;
          case 2 : if (c < minute) {
              c++;
              setPending();
            } else {
              setAvailable();
              c = 0;
            }
            break;
        }
      }
    }
    void writeColor() {
      digitalWrite(redPin, Red);
      digitalWrite(greenPin, Green);
      digitalWrite(bluePin, Blue);
    }
    void setColor(Color c) {
      Red = c.red;
      Green = c.green;
      Blue = c.blue;
      writeColor();
    }
    void setAvailable() {
      led_status = 0;
      setColor(color_green); 
    }
    void setOccupied() {
      led_status = 1;
      setColor(color_red) ; 
    }
    void setPending() {
      led_status = 2;
      setColor(color_blue);
    }
    void cancelPending() {
      setAvailable();
    }


};
//#define COMMON_ANODE
// id ,R ,G ,B, SensorPin
RGBled ParkingSlot1(1 , 2, 3, 4 , A0);
RGBled ParkingSlot2(2 , 5, 6, 7, A1);
RGBled ParkingSlot3(3 , 8, 9, 10 , A2);
RGBled ParkingSlot4(4 , 11, 12, 13, A4);
//RGBled green(  0,   255, 0, 0 , 11, 10, 9);
//RGBled yellow( 100, 200, 0, 0 , 11, 10, 9);
//Color white(   255,   255,   255, 0 , 9, 10, 11);
String statusToString() {
  String x =
    String("<") + ParkingSlot1.id + String(":") + ParkingSlot1.led_status + String(",") +
    ParkingSlot2.id + String(":") + ParkingSlot2.led_status + String(",") +
    ParkingSlot3.id + String(":") + ParkingSlot3.led_status + String(",") +
    ParkingSlot4.id + String(":") + ParkingSlot4.led_status +
    String(">");
  return x;
}
void setPendingID(char c) {
  switch (c) {
    case '1': ParkingSlot1.setPending() ; break;
    case '2': ParkingSlot2.setPending() ; break;
    case '3': ParkingSlot3.setPending(); break;
    case '4': ParkingSlot4.setPending(); break;
  }
}
void refresh () {
  ParkingSlot1.start();
  ParkingSlot2.start();
  ParkingSlot3.start();
  ParkingSlot4.start();
}
void setup() {
  //  Serial.begin(115200); // debugg
  Serial.begin(9600);
  while (!Serial) {
    ; // wait for serial port to connect. Needed for native USB port only
  }
  //Serial.println(minute);

}
String s = "";
String pending = "";
bool flag = false ;
//void loop (){
////       ParkingSlot1.setColor(color_blue);  ParkingSlot1.writeColor();
//ParkingSlot4.setAvailable();
//delay(1000);
//
//}
void loop() { // run over and over
  // refresh
  for (int i = 0 ; i <10;i++){
  refresh(); // 1
  delay(500);
  }
 
  
  String s = statusToString();
  for (int i = 0 ; i < s.length() ; i++) {
    Serial.write(s.charAt(i));
  }
  pending = "";
  while (Serial.available()) {
    char c = Serial.read();
    pending += c ;
  }
  if (!pending.startsWith("f")) {
    for (int i = 0 ; i < pending.length() ; i += 2) {
      char c = pending.charAt(i);
      setPendingID(c);
    }
  }

}
