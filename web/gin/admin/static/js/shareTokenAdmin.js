let checkboxes; // 全局变量
let shareTokenData

function renderShareTokenTable(data) {
    getDataFromAPI(serverUrl + "/api/v1/getAllShareToken").then(data => {
        let tableBody = document.querySelector('#shareTokenTable tbody');
        shareTokenData = data;
        data.forEach(token => {
            const row = document.createElement('tr');
            let date = new Date(token.UpdateTime);
            // 在原始时间的基础上加上600秒
            date.setSeconds(date.getSeconds() + parseInt(token.ExpiresTime));
            // 格式化为本地时间字符串
            let formattedTime = date.toLocaleString();
            if (token.ExpiresTime === 0) {
                formattedTime = "永久有效";
            } else if (token.ExpiresTime < 0) {
                formattedTime = "已过期";
            }
            row.innerHTML = `
            <th scope="row"><input class="form-check-input me-1" type="checkbox" value=""></th>
            <td>${token.ID}</td>
            <td>${token.UserID}</td>
            <td>${token.UniqueName}</td>
            <td>${formattedTime}</td>
            <td>${token.SiteLimit}</td>
            <td>${token.SK}</td>
            <td>${token.UpdateTime}</td>
            <td>${token.Comment}</td>
        `;
            tableBody.appendChild(row);
        });
    })
}

function bindSelectAllCheckbox() {
    const selectAllCheckbox = document.querySelector('#selectAllCheckbox');
    checkboxes = document.querySelectorAll('#shareTokenTable tbody input[type="checkbox"]');
    selectAllCheckbox.addEventListener('change', function () {
        checkboxes.forEach(checkbox => {
            checkbox.checked = selectAllCheckbox.checked;
        });
    });
}

function handleSubmitForm() {
    const form = document.querySelector('#addShareTokenForm');
    const submitBtn = document.querySelector('#submitBtn');

    submitBtn.addEventListener('click', function () {
        const formData = new FormData(form);
        const jsonData = {};

        for (let pair of formData.entries()) {
            jsonData[pair[0]] = pair[1];
        }
        console.log(jsonData);

        const url = 'https://example.com/api'; // 替换为您的目标地址
        const options = {
            method: 'POST', headers: {
                'Content-Type': 'application/json'
            }, body: JSON.stringify(jsonData)
        };

        fetch(url, options)
            .then(response => response.json())
            .then(data => {
                // 处理响应数据
                console.log(data);
            })
            .catch(error => {
                // 处理请求错误
                console.error(error);
            });
    });
}

function handleUpdateBtn() {
    // 监听更新按钮的点击事件
    const updateButton = document.querySelector('#updateButton');
    updateButton.addEventListener('click', function () {
        checkboxes = document.querySelectorAll('#shareTokenTable tbody input[type="checkbox"]');
        checkboxes.forEach((checkbox, index) => {
            if (checkbox.checked) {
                const row = checkbox.closest('tr');
                const rowData = shareTokenData[index];
                sendDataToGetShareToken(fakeopenUrl, rowData);
            }
        });
        // 在这里可以执行其他操作，例如将选中行的数据发送到服务器等
    });
}

async function getDataFromAPI(url) {
    try {
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        return await response.json();
    } catch (error) {
        console.error('Error:', error.message);
        return null;
    }
}

function getAccessTokenByUserID(userID) {
    let url = serverUrl + '/api/v1/getAccessToken?userID=' + encodeURIComponent(userID);
    console.log(url);
    return fetch(url)
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.text();
        })
        .catch(error => {
            console.error('Error:', error);
        })
}

function sendDataToGetShareToken(url, data) {

    getAccessTokenByUserID(data.UserID).then(accessToken => {
        var myHeaders = new Headers();
        myHeaders.append("Accept", "*/*");
        myHeaders.append("Cache-Control", "no-cache");
        myHeaders.append("Host", "ai.fakeopen.com");
        myHeaders.append("Connection", "keep-alive");
        myHeaders.append("Content-Type", "application/x-www-form-urlencoded");
        let urlencoded = new URLSearchParams();
        urlencoded.append("unique_name", data.UniqueName);
        urlencoded.append("access_token", accessToken);
        urlencoded.append("expires_in", data.ExpiresTime);
        urlencoded.append("site_limit", data.SiteLimit);
        console.log(accessToken);
        let requestOptions = {
            headers: myHeaders,
            method: 'POST',
            body: urlencoded,
            redirect: 'follow',
            mode:'cors'
        };

        return fetch(url, requestOptions)
            .then(response => response.json())
            .then * (jsonData => {
            console.log(jsonData);
            return jsonData;
        })
            .catch(error => {
                console.error(error);
            })
    })
}

renderShareTokenTable();// 显示所有来自于json的数据
bindSelectAllCheckbox();
handleSubmitForm(); // 提交表单
handleUpdateBtn();