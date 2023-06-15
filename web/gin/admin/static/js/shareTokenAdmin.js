let checkboxes; // 全局变量

function renderShareTokenTable(data) {
    getDataFromAPI().then(data => {
        let tableBody = document.querySelector('#shareTokenTable tbody');
        data.forEach(token => {
            const row = document.createElement('tr');
            let goTime = new Date(token.ExpiresTime);
            goTime.setSeconds(goTime.getSeconds() + 600);
            console.log(goTime.toLocaleString());
            row.innerHTML = `
            <th scope="row"><input class="form-check-input me-1" type="checkbox" value=""></th>
            <td>${token.ID}</td>
            <td>${token.UserID}</td>
            <td>${token.UniqueName}</td>
            <td>${goTime.toLocaleString()}</td>
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
        const selectedRows = [];
        checkboxes = document.querySelectorAll('#shareTokenTable tbody input[type="checkbox"]');
        checkboxes.forEach((checkbox, index) => {
            if (checkbox.checked) {
                const row = checkbox.closest('tr');
                const rowData = shareTokenData[index];
                selectedRows.push(rowData);
            }
        });

        console.log('选中的行数据:', selectedRows);
        sendDataToGetShareToken("https://example.com/api", selectedRows);
        // 在这里可以执行其他操作，例如将选中行的数据发送到服务器等
    });
}

async function getDataFromAPI(url) {
    // todo))这里是构建一组虚拟的数据，发布之后把它删掉
    return shareTokenData;
    try {
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        const data = await response.json();
        return data;
    } catch (error) {
        console.error('Error:', error.message);
        return null;
    }
}

function sendDataToGetShareToken(url, data) {
    let myHeaders = new Headers();
    myHeaders.append("User-Agent", navigator.userAgent);
    myHeaders.append("Accept", "*/*");
    myHeaders.append("Cache-Control", "no-cache");
    myHeaders.append("Host", "ai.fakeopen.com");
    myHeaders.append("Connection", "keep-alive");
    myHeaders.append("Content-Type", "application/x-www-form-urlencoded");

    let urlencoded = new URLSearchParams();
    urlencoded.append("unique_name", data.UniqueName);
    urlencoded.append("access_token", data.AssessToken);
    urlencoded.append("expires_in", data.ExpiresTime);
    urlencoded.append("site_limit", data.SiteLimit);

    let requestOptions = {
        method: 'POST',
        headers: myHeaders,
        body: urlencoded,
        redirect: 'follow'
    };

    return fetch(url, requestOptions)
        .then(response => response.text())
        .then(result => {
            console.log(result);
            return result;
        })
        .catch(error => {
            console.log('error', error);
            throw error;
        });
}

renderShareTokenTable();// 显示所有来自于json的数据
document.addEventListener('DOMContentLoaded', function () {
    console.log("dom loaded");
    bindSelectAllCheckbox();
    handleSubmitForm(); // 提交表单
    handleUpdateBtn();
});