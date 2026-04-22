<template>
    <div class="page-install">
        <ErrorPage
                code="500"
                desc="抱歉，系统尚未安装或配置异常，请联系管理员"
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