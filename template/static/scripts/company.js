const html = document.documentElement;

if (localStorage.getItem("theme") === "dark") {
    html.setAttribute("data-theme", "dark");
}