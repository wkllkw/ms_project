<template>
    <div class="page-500">
        <ErrorPage
                code="500"
                desc="抱歉，服务器出错了"
                subDesc="服务器开小差了，请稍后重试或联系系统管理员"
        >
            <img slot="img" src="../../assets/image/error/500.svg" alt="">
        </ErrorPage>
    </div>
</template>
<script>
    import ErrorPage from '../../components/error/errorPage';
    export default {
        components: {
            ErrorPage
        }
    }
</script>
