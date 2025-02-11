const themeIcon = document.getElementById("theme-toggler");
const html = document.documentElement;

if (localStorage.getItem("theme") === "dark") {
    html.setAttribute("data-theme", "dark");
    themeIcon.classList.replace("fa-moon", "fa-sun");
}

themeIcon.addEventListener("click", () => {
    const currentTheme = html.getAttribute("data-theme");
    const newTheme = currentTheme === "dark" ? "light" : "dark";

    html.setAttribute("data-theme", newTheme);
    localStorage.setItem("theme", newTheme);

    themeIcon.classList.toggle("fa-moon");
    themeIcon.classList.toggle("fa-sun");

});
