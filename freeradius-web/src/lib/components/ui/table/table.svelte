<script lang="ts">
	import { cn } from '$lib/utils.js';
	let klass: string | undefined = undefined;
	export { klass as class };
</script>

<div class={cn('relative w-full overflow-auto', klass)} {...$$restProps}>
	<table class="w-full caption-bottom text-sm">
		<slot />
	</table>
</div>
