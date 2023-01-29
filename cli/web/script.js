const requestsEl = document.getElementById("requests");
const infoSections = document.getElementsByClassName("card-info");
let requests = [];

for (let infoSection of infoSections) {
    let title = infoSection.getElementsByClassName("header-title")[0];

    title.addEventListener("click", () => {
        let is_open = !(infoSection.dataset.isOpen === "true");
        infoSection.dataset.isOpen = is_open;

        let detailsEl = infoSection.getElementsByClassName("details")[0];
        let arrowSvg = infoSection.getElementsByTagName("svg")[0];
        if (is_open) {
            detailsEl.classList.remove("hidden");
            arrowSvg.classList.add("open");
        } else {
            detailsEl.classList.add("hidden");
            arrowSvg.classList.remove("open");
        }
    });
}

function createElementFromHTML(htmlString) {
    var div = document.createElement("div");
    div.innerHTML = htmlString.trim();
    return div.firstChild;
}

const getMethodColor = (method) => {
    colors = {
        GET: "sky-500",
        POST: "green-500",
        PUT: "orange-500",
        PATCH: "emerald-500",
        DELETE: "rose-500",
        HEAD: "indigo-500",
        OPTIONS: "blue-500",
        DEFAULT: "gray-500",
    };
    return method in colors ? colors[method] : colors["DEFAULT"];
};

const getStatusColor = (status) => {
    let first_digit = Math.floor(status / 100);
    colors = {
        1: "gray-500",
        2: "green-500",
        3: "yellow-500",
        4: "red-500",
        5: "rose-500",
        DEFAULT: "gray-500",
    };
    return first_digit in colors ? colors[first_digit] : colors["DEFAULT"];
};

const addRequest = (request) => {
    let methodColor = getMethodColor(request.method);
    const requestElHTML = `
    <div class="card cursor-pointer request border-l border-t" onclick="selectRequest(${
        request.id
    })" data-is-active="false" data-id=${request.id}>
        <div class="method w-20 text-${methodColor}">${request.method}</div>
        <div class="path flex-1 text-black/60" title=${request.url}>
		${request.url.slice(0, 20)}${request.url.length > 20 ? "..." : ""}
		</div>
        <div class="status w-12 text-right"><div class="loader"></div></div>
    </div>
    `;
    const requestElJS = createElementFromHTML(requestElHTML);
    requestsEl.prepend(requestElJS);
};

const update_request_status = (request_id, status) => {
    let requestEl = document.querySelector(`[data-id='${request_id}']`);
    let statusEl = requestEl.querySelector(".status");
    let statusColor = getStatusColor(status);

    statusEl.classList.add(`text-${statusColor}`);
    statusEl.innerHTML = status;
};
const prettifyJson = (json_str) => {
    json = JSON.parse(json_str);
    return JSON.stringify(json, null, 2);
};

const selectRequest = (request_id) => {
    request = requests.find((request) => request.id === request_id);
    let requestEls = document.querySelector("#requests");
    for (let requestEl of requestEls.childNodes) {
        if (parseInt(requestEl.dataset.id) === request_id) {
            requestEl.dataset.isActive = true;
        } else {
            requestEl.dataset.isActive = false;
        }
    }

    // console.log(requestEl);
};

const changeRequestInfo = (request) => {
    let infoSectionEl = document.getElementById("info");
    let requestBodyEl = infoSectionEl
        .querySelector('[data-title="requestBody"]')
        .querySelector(".details");
    let responseBodyEl = infoSectionEl
        .querySelector('[data-title="responseBody"]')
        .querySelector(".details");
    let requestHeadersEl = infoSectionEl
        .querySelector('[data-title="requestHeaders"]')
        .querySelector(".details");
    let responseHeadersEl = infoSectionEl
        .querySelector('[data-title="responseHeaders"]')
        .querySelector(".details");

    let dummyJSON = `[
        {
          "title": "apples",
          "count": [12000, 20000],
          "description": {"text": "...", "sensitive": false}
        },
        {
          "title": "oranges",
          "count": [17500, null],
          "description": {"text": "...", "sensitive": false}
        }
      ]`;

    requestBodyEl.innerHTML = `<pre>
            <code class="language-json text-normal">${prettifyJson(
                dummyJSON
            )}</code>
        </pre>`;
};

changeRequestInfo();

const handleEvents = async () => {
    for (let i = 0; i < 10; i++) {
        await new Promise((r) => setTimeout(r, Math.random() * 1000));

        handleEvent(
            `{"data": {"id": ${i}, "method":"GET","url":"https://google.com","body":"","headers":{"keep-alive":"true"}}}`
        );
        await new Promise((r) => setTimeout(r, 1500));

        setTimeout(() => {}, 3000);
        handleEvent(
            `{"data": {"request_id": ${i}, "status":200,"headers":{"accept":"Json"},"body":[]}}`
        );
    }
};

const handleEvent = (event_str) => {
    let event = JSON.parse(event_str)["data"];
    if ("request_id" in event) {
        // Event is response
        request = requests.find((request) => request.id === event.request_id);
        if (request) {
            request.response = event;
            update_request_status(request.id, event.status);
        }
    } else {
        try {
            event["response"] = {};
            requests.push(event);
            addRequest(event);
        } catch {
            console.log("Could not load request");
        }
    }
};

function main() {
    handleEvents();
}

main();
