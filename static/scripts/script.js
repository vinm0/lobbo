
// Set copyright year
const copyright = document.getElementById("year");
let d = new Date();
copyright.innerHTML = d.getFullYear();

function highlightCurrentNav(elementID) {
    nav = document.getElementById(elementID)
    nav.style.boxShadow = "inset 0 0.3rem 0 0.1rem rgb(104, 237, 255)"
}