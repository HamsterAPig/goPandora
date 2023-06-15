function renderShareTokenTable(data) {
    getDataFromAPI().then(data => {
        let tableBody = document.querySelector('#shareTokenTable tbody');
        data.forEach(token => {
            const row = document.createElement('tr');
            row.innerHTML = `
            <th scope="row"><input class="form-check-input me-1" type="checkbox" value=""></th>
            <td>${token.ID}</td>
            <td>${token.UserID}</td>
            <td>${token.UniqueName}</td>
            <td>${token.ExpiresAt}</td>
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
    const checkboxes = document.querySelectorAll('#shareTokenTable tbody input[type="checkbox"]');

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

renderShareTokenTable();// 显示所有来自于json的数据
bindSelectAllCheckbox();// 监听复选框
handleSubmitForm(); // 提交表单
