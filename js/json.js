function browserSupportsDateInput() {
  var b = false; 
  try {
    var tester = document.createElement('input');
    tester.type = "date";
    console.log("tester.type=" + tester.type);
    console.log("browser supports date input?" + (tester.type === "date"));
    b = tester.type === "date"; 
  } catch (e) {
    console.log("oops, got error testing browser date input support" + e);
  } finally {return b;}
}
