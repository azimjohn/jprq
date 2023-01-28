const requestsEl = document.getElementById("requests");

function MakeRequest(
    method,
    path,
    statusCode,
    generalInfo,
    requestHeaders,
    responseHeaders
) {
    return {
        method,
        path,
        statusCode,
        generalInfo,
        requestHeaders,
        responseHeaders,
    };
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
    let statusColor = getStatusColor(request.statusCode);
    const requestElHTML = `
    <div class="card cursor-pointer">
        <div class="method w-20 text-${methodColor}">${request.method}</div>
        <div class="path flex-1 text-black/60" title=${request.path}>
		${request.path.slice(0, 20)}${request.path.length > 20 ? "..." : ""}
		</div>
        <div class="status w-12 text-${statusColor} text-center">${
        request.statusCode
    }</div>
    </div>
    `;
    const requestElJS = createElementFromHTML(requestElHTML);
    requestsEl.appendChild(requestElJS);
};

addRequest(MakeRequest("GET", "/posts", 200, "", "", ""));
addRequest(MakeRequest("POST", "/posts", 201, "", "", ""));
addRequest(MakeRequest("DELETE", "/posts", 202, "", "", ""));
addRequest(MakeRequest("OPTIONS", "/posts", 200, "", "", ""));
addRequest(MakeRequest("PUT", "/posts", 204, "", "", ""));
addRequest(MakeRequest("PATCH", "/posts", 404, "", "", ""));
addRequest(
    MakeRequest("PATCH", "/posts?adflkjadkfjdsjklflkjb", 504, "", "", "")
);
addRequest(MakeRequest("HEAD", "/posts", 304, "", "", ""));
addRequest(MakeRequest("HEAD", "/posts", 304, "", "", ""));
addRequest(MakeRequest("HEAD", "/posts", 304, "", "", ""));
addRequest(MakeRequest("HEAD", "/posts", 304, "", "", ""));
addRequest(MakeRequest("HEAD", "/posts", 304, "", "", ""));
addRequest(MakeRequest("HEAD", "/posts", 304, "", "", ""));
