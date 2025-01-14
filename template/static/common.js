

// neeeded data, do not change
window.alldata = [];
window.savedPages = {};
window.searched = [];

window.dlen = 0;
window.pageSize = 0;
window.lastIndex = 0;
window.currPage = 0;   

window.dataList = document.getElementById('dataList');
window.filter = document.getElementById('appType');
window.prevbtn = document.getElementById('prev-btn');
window.nextbtn = document.getElementById('next-btn');

window.cardTemplate = ``;

// function to filter results 
// only one filter can be used for now
function filterData() {
    let query = filter.value;

    if (query == "All") {
        searched = alldata;
    } else {
        const result = alldata.filter(val => val.Type.includes(query));
        searched = result;
    }
    middle();
}
// to search 
// searches for match in all fields of all objects
// no specific field matching for now
function search() {
    const query = searchbar.value.toLowerCase();
    if (query == "") {
        searched = alldata;
        middle();
        return;
    }
    const result = alldata.filter(job =>
    Object.values(job).some(val => val.toString().toLowerCase().includes(query)));
    searched = result;
    middle();
}

// shows the 'loading' spinner
window.showSpinner = function () {
    document.getElementById('spinner').style.display = 'block'; 
};
// hides the 'loading' spinner
function hideSpinner() {
    document.getElementById('spinner').style.display = 'none'; 
}

// lists the cards one by one for the given pageSize
function listjobs() {
    if (searched.length != 0) {
        let i = 0;
        for (i = lastIndex; i < dlen; i++) {
            if (pageSize == 0) {
                break;
            }

            let obj = searched[i]; 
            const card = cardTemplate;
            dataList.innerHTML += card;
            pageSize--;
        }
        lastIndex = i;
    }
}

// houses the next and prev button actions
// triggers listdata() when needed
// or applies the saved page is present in savedPages{}
function pagination() {

    if (searched != null) {
        dlen = searched.length;
    }

    prevbtn.onclick = function () {
        currPage -= 1;
        dataList.innerHTML = savedPages[currPage].Template;
        pageSize = 10 - savedPages[currPage].Length;
        displaybtn();
    }

    nextbtn.onclick = function () {
        currPage += 1;
        pageSize = 10; // this defines the number of cards per page
        
        if (savedPages[currPage] != null)  {
            dataList.innerHTML = savedPages[currPage].Template;
            pageSize = 10 - savedPages[currPage].Length;
        } else {
            dataList.innerHTML = ``;

            if (searched != null && dlen > lastIndex) {
                listjobs();
                savedPages[currPage] = {
                    Template: dataList.innerHTML,
                    Length: (10 - pageSize)
                }
            }
        }
        displaybtn();
    }

}

// displays appropriate next/prev buttons
function displaybtn() {
    prevbtn.style.visibility = currPage <= 1 ? 'hidden' : 'visible';

    if (pageSize > 0) {
        nextbtn.style.visibility = 'hidden';
        dataList.innerHTML += `
            <div class="details">
                Thats all for now !
            </div>
        `;
    } else {
        nextbtn.style.visibility = 'visible';
    }
}

// acts as a converging point for the pagination functions
function middle() {
    pagination();
    lastIndex = currPage = 0;
    savedPages = {};
    nextbtn.click();
}

// starts a listener for the search bar, triggers search() every time a character is changed 
const searchbar = document.getElementById('searchBox');
searchbar.addEventListener('input', search);


console.log("first");